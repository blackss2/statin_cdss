package reservation

import (
	"net/url"
	"time"

	"git.tnicdev.co.kr/hami/CTC_Web/pkg/store"
	_ "git.tnicdev.co.kr/hami/CTC_Web/pkg/store/driver/rethinkdb"
)

type Reservation struct {
	Id       string                `json:"id,omitempty"`
	Name     string                `json:"name"`
	StudyId  string                `json:"study_id"`
	TDate    time.Time             `json:"t_date"`
	Subjects []*ReservationSubject `json:"subjects"`
	TCreate  time.Time             `json:"t_create"`
	ActorId  string                `json:"actor_id"`
}

type ReservationSubject struct {
	SubjectId string `json:"subject_id"`
	Minutes   int64  `json:"minutes"`
	Status    int64  `json:"status"`
}

type Store struct {
	store.Store
}

func NewStore(u *url.URL, tableName string) (*Store, error) {
	s, err := store.NewStore(u.Scheme)
	if err != nil {
		return nil, err
	}
	err = s.Connect(u.Host + u.Path)
	if err != nil {
		return nil, err
	}
	err = s.InitTable(store.TableOption{
		TableName:   tableName,
		TableCreate: true,
		IndexNames:  []string{"study_id", "t_create"},
		IndexCreate: true,
		IndexDelete: true,
	})
	if err != nil {
		return nil, err
	}

	st := &Store{
		Store: s,
	}
	return st, nil
}

func (st *Store) Insert(StudyId string, Name string, TDate time.Time, t_create time.Time, ActorId string) (string, error) {
	item := &Reservation{
		StudyId:  StudyId,
		Name:     Name,
		TDate:    TDate,
		Subjects: make([]*ReservationSubject, 0),
		TCreate:  t_create,
		ActorId:  ActorId,
	}

	id, err := st.Store.Insert(item)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (st *Store) AddSubject(id string, SubjectId string, Minutes int64, Status int64) error {
	var item Reservation
	err := st.Get(id, &item)
	if err != nil {
		return err
	}

	for _, s := range item.Subjects {
		if s.SubjectId == SubjectId {
			return ErrExistReservationSubject
		}
	}

	item.Subjects = append(item.Subjects, &ReservationSubject{
		SubjectId: SubjectId,
		Minutes:   Minutes,
		Status:    Status,
	})
	return st.Store.Update(id, &item)
}

func (st *Store) Update(id string, Name string) error {
	var item Reservation
	err := st.Get(id, &item)
	if err != nil {
		return err
	}

	isChanged := false
	if item.Name != Name {
		item.Name = Name
		isChanged = true
	}

	if isChanged {
		return st.Store.Update(id, &item)
	} else {
		return nil
	}
}

func (st *Store) UpdateSubject(id string, SubjectId string, Minutes int64, Status int64) error {
	var item Reservation
	err := st.Get(id, &item)
	if err != nil {
		return err
	}

	var subject *ReservationSubject
	for _, s := range item.Subjects {
		if s.SubjectId == SubjectId {
			subject = s
		}
	}

	if subject == nil {
		return ErrNotExistReservationSubject
	}

	subject.Minutes = Minutes
	subject.Status = Status
	return st.Store.Update(id, &item)
}

func (st *Store) Delete(id string) error {
	return st.Store.Delete(id)
}

func (st *Store) DeleteSubject(id string, SubjectId string) error {
	var item Reservation
	err := st.Get(id, &item)
	if err != nil {
		return err
	}

	idx := -1
	for i, s := range item.Subjects {
		if s.SubjectId == SubjectId {
			idx = i
			break
		}
	}

	if idx < 0 {
		return ErrNotExistReservationSubject
	}

	if idx == 0 {
		item.Subjects = item.Subjects[1:]
	} else if idx == len(item.Subjects)-1 {
		item.Subjects = item.Subjects[:len(item.Subjects)-1]
	} else {
		item.Subjects = append(item.Subjects[:idx], item.Subjects[idx+1:]...)
	}
	return st.Store.Update(id, &item)
}
