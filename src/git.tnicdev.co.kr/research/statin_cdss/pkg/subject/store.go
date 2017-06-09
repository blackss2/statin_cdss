package subject

import (
	"net/url"
	"time"

	"git.tnicdev.co.kr/research/statin_cdss/pkg/store"
	_ "git.tnicdev.co.kr/research/statin_cdss/pkg/store/driver/rethinkdb"
)

type Subject struct {
	Id        string    `json:"id,omitempty"`
	SubjectId string    `json:"subject_id"`
	Datas     []*Data   `json:"datas"`
	Share     bool      `json:"share"`
	OwnerId   string    `json:"owner_id"`
	TCreate   time.Time `json:"t_create"`
}

type Estimation struct {
	DangerousGroup string  `json:"dangerous_group"`
	TargetLDL      float64 `json:"target_ldl"`
}

type Prescription struct {
	Statins []string `json:"statins"`
	Levels  []string `json:"levels"`
}

type Data struct {
	Demography     Demography     `json:"demography"`
	BloodPressure  BloodPressure  `json:"blood_pressure"`
	StatinFirst    StatinFirst    `json:"statin_first"`
	StatinsLast    StatinsLast    `json:"statin_last"`
	BloodTest      BloodTest      `json:"blood_test"`
	MedicalHistory MedicalHistory `json:"medical_history"`
	FamilyHistory  FamilyHistory  `json:"family_history"`
	Estimation     Estimation     `json:"estimation"`
	Prescription   Prescription   `json:"prescription"`
	TCreate        time.Time      `json:"t_create"`
}

type Demography struct {
	BirthDate string  `json:"birth_date"`
	Age       int64   `json:"age"`
	Sex       string  `json:"sex"`
	Height    float64 `json:"height"`
	Weight    float64 `json:"weight"`
}

type BloodPressure struct {
	Date      string `json:"date"`
	Systolic  int64  `json:"systolic"`
	Diastolic int64  `json:"diastolic"`
}

type StatinFirst struct {
	Dept   string `json:"dept"`
	Code   string `json:"code"`
	Date   string `json:"date"`
	Period int64  `json:"period"`
}

type StatinsLast struct {
	Dept   string `json:"dept"`
	Code   string `json:"code"`
	Date   string `json:"date"`
	Period int64  `json:"period"`
}

type BloodTest struct {
	Date             string  `json:"date"`
	LDL              float64 `json:"ldl_c"`
	HDL              float64 `json:"hdl"`
	TotalCholesterol float64 `json:"total_cholesterol"`
	Glucose          float64 `json:"glucose"`
}

type MedicalHistory struct {
	TransientStroke    bool `json:"transient_stroke"`
	PeripheralVascular bool `json:"peripheral_vascular"`
	Carotid            bool `json:"carotid"`
	AbdominalAneurysm  bool `json:"abdominal_aneurysm"`
	Diabetes           bool `json:"diabetes"`
	CoronaryArtery     bool `json:"coronary_artery"`
	IschemicStroke     bool `json:"ischemic_stroke"`
	HighBloodPressure  bool `json:"high_blood_pressure"`
	Smoking            bool `json:"smoking"`
}

type FamilyHistory struct {
	CoronaryArtery bool `json:"coronary_artery"`
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
		IndexNames:  []string{"owner_id", "share", "t_create"},
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
				IndexByValues: []interface{}{OwnerId},
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
				IndexBy:       "owner_id",
				IndexByValues: []interface{}{OwnerId},
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

func (st *Store) ListShare() ([]*Subject, error) {
	opts := []store.ListOption{
		store.ListOption{
			WhereOption: store.WhereOption{
				IndexBy:       "share",
				IndexByValues: []interface{}{true},
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
		Datas:     make([]*Data, 0),
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

func (st *Store) Update() {
	panic("not support update")
}

func (st *Store) AppendData(id string, Data *Data) error {
	var item Subject
	err := st.Get(id, &item)
	if err != nil {
		return err
	}

	item.Datas = append(item.Datas, Data)
	return st.Store.Update(id, &item)
}

func (st *Store) UpdateLastData(id string, Data *Data) error {
	var item Subject
	err := st.Get(id, &item)
	if err != nil {
		return err
	}

	if len(item.Datas) == 0 {
		return ErrNotExistSubjectData
	}
	item.Datas[len(item.Datas)-1] = Data
	return st.Store.Update(id, &item)
}

func (st *Store) SetShare(id string, Share bool) error {
	var item Subject
	err := st.Get(id, &item)
	if err != nil {
		return err
	}

	item.Share = Share
	return st.Store.Update(id, &item)
}

func (st *Store) Delete(id string) error {
	return st.Store.Delete(id)
}
