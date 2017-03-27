package user

import (
	"net/url"
	"time"

	"git.tnicdev.co.kr/hami/CTC_Web/pkg/store"
	_ "git.tnicdev.co.kr/hami/CTC_Web/pkg/store/driver/rethinkdb"
)

type User struct {
	Id           string    `json:"id,omitempty"`
	UserId       string    `json:"userid"`
	Password     string    `json:"password"`
	Name         string    `json:"name"`
	Birth        string    `json:"birth"`
	Mobile       string    `json:"mobile"`
	Organization string    `json:"organization"`
	Position     string    `json:"position"`
	Role         string    `json:"role"`
	Disable      bool      `json:"disable"`
	TCreate      time.Time `json:"t_create"`
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

func (st *Store) GetByUserId(userid string, CheckEnable bool) (*User, error) {
	opts := []store.ListOption{
		store.ListOption{
			WhereOption: store.WhereOption{
				FieldBy:       "userid",
				FieldByValues: []interface{}{userid},
			},
		},
	}

	if CheckEnable {
		opts = append(opts, store.ListOption{
			WhereOption: store.WhereOption{
				FieldBy:       "disable",
				FieldByValues: []interface{}{false},
			},
		})
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

func (st *Store) ListByRole(Role string, CheckEnable bool) ([]*User, error) {
	opts := []store.ListOption{
		store.ListOption{
			WhereOption: store.WhereOption{
				FieldBy:       "role",
				FieldByValues: []interface{}{Role},
			},
		},
	}

	if CheckEnable {
		opts = append(opts, store.ListOption{
			WhereOption: store.WhereOption{
				FieldBy:       "disable",
				FieldByValues: []interface{}{false},
			},
		})
	}

	var list []*User
	err := st.List(&list, opts...)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (st *Store) Insert(userid string, password string, name string, birth string, mobile string, organization string, position string, role string, t_create time.Time) (string, error) {
	item := &User{
		UserId:       userid,
		Password:     password,
		Name:         name,
		Birth:        birth,
		Mobile:       mobile,
		Organization: organization,
		Position:     position,
		Role:         role,
		Disable:      false,
		TCreate:      t_create,
	}

	_, err := st.GetByUserId(userid, false)
	if err == nil {
		return "", ErrExistUserId
	}

	id, err := st.Store.Insert(item)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (st *Store) Update(id string, name string, birth string, mobile string, organization string, position string) error {
	var item User
	err := st.Get(id, &item)
	if err != nil {
		return err
	}

	isChanged := false
	if len(name) > 0 && item.Name != name {
		item.Name = name
		isChanged = true
	}
	if len(birth) > 0 && item.Birth != birth {
		item.Birth = birth
		isChanged = true
	}
	if len(mobile) > 0 && item.Mobile != mobile {
		item.Mobile = mobile
		isChanged = true
	}
	if len(organization) > 0 && item.Organization != organization {
		item.Organization = organization
		isChanged = true
	}
	if len(position) > 0 && item.Position != position {
		item.Position = position
		isChanged = true
	}

	if isChanged {
		return st.Store.Update(id, &item)
	} else {
		return nil
	}
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

func (st *Store) UpdateRole(id string, role string) error {
	var item User
	err := st.Get(id, &item)
	if err != nil {
		return err
	}

	isChanged := false
	if len(role) > 0 && item.Role != role {
		item.Role = role
		isChanged = true
	}

	if isChanged {
		return st.Store.Update(id, &item)
	} else {
		return nil
	}
}

func (st *Store) SetDisable(id string, disable bool) error {
	var item User
	err := st.Get(id, &item)
	if err != nil {
		return err
	}

	isChanged := false
	if item.Disable != disable {
		item.Disable = disable
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
