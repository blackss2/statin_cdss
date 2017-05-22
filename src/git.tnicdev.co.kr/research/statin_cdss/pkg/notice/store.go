package notice

import (
	"net/url"
	"time"

	"git.tnicdev.co.kr/research/statin_cdss/pkg/store"
	_ "git.tnicdev.co.kr/research/statin_cdss/pkg/store/driver/rethinkdb"
)

type Notice struct {
	Id      string    `json:"id,omitempty"`
	StudyId string    `json:"study_id"`
	Content string    `json:"content"`
	TCreate time.Time `json:"t_create"`
	ActorId string    `json:"actor_id"`
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

func (st *Store) Insert(StudyId string, Content string, t_create time.Time, ActorId string) (string, error) {
	item := &Notice{
		StudyId: StudyId,
		Content: Content,
		TCreate: t_create,
		ActorId: ActorId,
	}

	id, err := st.Store.Insert(item)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (st *Store) Update(id string, Content string) error {
	var item Notice
	err := st.Get(id, &item)
	if err != nil {
		return err
	}

	isChanged := false
	if len(Content) > 0 && item.Content != Content {
		item.Content = Content
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
