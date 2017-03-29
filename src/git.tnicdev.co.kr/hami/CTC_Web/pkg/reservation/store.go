package reservation

import (
	"net/url"
	"time"

	"git.tnicdev.co.kr/hami/CTC_Web/pkg/store"
	_ "git.tnicdev.co.kr/hami/CTC_Web/pkg/store/driver/rethinkdb"
)

type Reservation struct {
	Id       string    `json:"id,omitempty"`
	StudyId  string    `json:"study_id"`
	TDate    time.Time `json:"t_date"`
	Subjects []*ReservationSubject
	TCreate  time.Time `json:"t_create"`
	ActorId  string    `json:"actor_id"`
}

type ReservationSubject struct {
	SubjectId    string    `json:"subject_id"`
	TReservation time.Time `json:"t_reservation"`
	Status       string    `json:"status"`
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

func (st *Store) Insert(StudyId string, TDate time.Time, t_create time.Time, ActorId string) (string, error) {
	item := &Reservation{
		StudyId:  StudyId,
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

func (st *Store) Update(id string, TDate time.Time) error {
	var item Reservation
	err := st.Get(id, &item)
	if err != nil {
		return err
	}

	isChanged := false
	if !item.TDate.Equal(TDate) {
		item.TDate = TDate
		isChanged = true
	}

	if isChanged {
		return st.Store.Update(id, &item)
	} else {
		return nil
	}
}

func (st *Store) Delete(id string) error {
	return st.Store.Delete(id)
}
