package subject

import (
	"net/url"
	"time"

	"git.tnicdev.co.kr/research/statin_cdss/pkg/store"
	_ "git.tnicdev.co.kr/research/statin_cdss/pkg/store/driver/rethinkdb"
)

type Subject struct {
	Id        string `json:"id,omitempty" gorethink:"id,omitempty"`
	SubjectId string `json:"subject_id" gorethink:"subject_id"`
	//TODO
	OwnerId string    `json:"owner_id" gorethink:"owner_id"`
	TCreate time.Time `json:"t_create" gorethink:"t_create"`
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
		IndexNames:  []string{"owner_id", "t_create"},
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

func (st *Store) GetBySubjectId(SubjectId string, OwnerId string) (*Subject, error) {
	opts := []store.ListOption{
		store.ListOption{
			WhereOption: store.WhereOption{
				IndexBy:       "owner_id",
				IndexByValues: []interface{}{SubjectId},
				FieldBy:       "subject_id",
				FieldByValues: []interface{}{SubjectId},
			},
		},
	}

	opts = append(opts, store.ListOption{
		Limit: 1,
	})

	var list []*Subject
	err := st.List(&list, opts...)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, ErrNotExistSubject
	}
	return list[0], nil
}

func (st *Store) ListByOwnerId(OwnerId string) ([]*Subject, error) {
	opts := []store.ListOption{
		store.ListOption{
			WhereOption: store.WhereOption{
				FieldBy:       "owner_id",
				FieldByValues: []interface{}{OwnerId},
			},
		},
	}

	var list []*Subject
	err := st.List(&list, opts...)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (st *Store) Insert(SubjectId string, OwnerId string, TCreate time.Time) (string, error) {
	item := &Subject{
		SubjectId: SubjectId,
		OwnerId:   OwnerId,
		TCreate:   TCreate,
	}

	_, err := st.GetBySubjectId(SubjectId, OwnerId)
	if err == nil {
		return "", ErrExistSubjectId
	}

	id, err := st.Store.Insert(item)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (st *Store) Update(id string, SubjectId string) error {
	var item Subject
	err := st.Get(id, &item)
	if err != nil {
		return err
	}

	isChanged := false
	if len(SubjectId) > 0 && item.SubjectId != SubjectId {
		item.SubjectId = SubjectId
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
