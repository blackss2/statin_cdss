package user

import (
	"net/url"
	"time"

	"git.tnicdev.co.kr/research/statin_cdss/pkg/store"
	_ "git.tnicdev.co.kr/research/statin_cdss/pkg/store/driver/rethinkdb"
)

type User struct {
	Id       string    `json:"id,omitempty"`
	UserId   string    `json:"userid"`
	Password string    `json:"password"`
	TCreate  time.Time `json:"t_create"`
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
		IndexNames:  []string{"t_create"},
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

func (st *Store) GetByUserId(userid string) (*User, error) {
	opts := []store.ListOption{
		store.ListOption{
			WhereOption: store.WhereOption{
				FieldBy:       "userid",
				FieldByValues: []interface{}{userid},
			},
		},
	}

	opts = append(opts, store.ListOption{
		Limit: 1,
	})

	var list []*User
	err := st.List(&list, opts...)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, ErrNotExistUser
	}
	return list[0], nil
}

func (st *Store) Insert(userid string, password string, t_create time.Time) (string, error) {
	item := &User{
		UserId:   userid,
		Password: password,
		TCreate:  t_create,
	}

	_, err := st.GetByUserId(userid)
	if err == nil {
		return "", ErrExistUserId
	}

	id, err := st.Store.Insert(item)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (st *Store) Update() {
	panic("not support update")
}

func (st *Store) UpdatePassword(id string, password string) error {
	var item User
	err := st.Get(id, &item)
	if err != nil {
		return err
	}

	isChanged := false
	if len(password) > 0 && item.Password != password {
		item.Password = password
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
