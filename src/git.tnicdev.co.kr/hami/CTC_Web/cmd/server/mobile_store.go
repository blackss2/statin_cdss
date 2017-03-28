package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/blackss2/utility/convert"
	"github.com/blackss2/utility/htmlwriter"
	"github.com/dancannon/gorethink"
)

type Study struct {
	Id      string    `json:"id,omitempty" gorethink:"id,omitempty"`
	Name    string    `json:"name" gorethink:"name"`
	TCreate time.Time `json:"t_create" gorethink:"t_create"`
	ActorId string    `json:"actor_id" gorethink:"actor_id"`
}

type StudyTable struct {
	DBTable
}

func NewStudyTable(session *gorethink.Session, db gorethink.Term) *StudyTable {
	st := &StudyTable{
		DBTable: DBTable{
			session: session,
			db:      db,
		},
	}
	st.Init("study")
	return st
}

func (st *StudyTable) Insert(Name string, TCreate time.Time, ActorId string) (*Study, error) {
	study := &Study{
		Name:    Name,
		TCreate: TCreate,
		ActorId: ActorId,
	}
	id, err := st.DBTable.Insert(study)
	if err != nil {
		return nil, err
	}
	study.Id = id
	return study, nil
}

func (st *StudyTable) Study(StudyId string) (*Study, error) {
	res, err := st.table.Get(StudyId).Run(st.session)
	if err != nil {
		return nil, err
	}

	if res.IsNil() {
		return nil, fmt.Errorf("not found")
	} else {
		study := &Study{}
		res.One(&study)

		return study, nil
	}
}

func (st *StudyTable) List() ([]*Study, error) {
	res, err := st.table.Run(st.session)
	if err != nil {
		return nil, err
	}

	list := make([]*Study, 0)
	res.All(&list)

	return list, nil
}

type Subject struct {
	Id        string    `json:"id,omitempty" gorethink:"id,omitempty"`
	Name      string    `json:"name" gorethink:"name"`
	ScrNo     string    `json:"scrno" gorethink:"scrno"`
	Sex       string    `json:"sex" gorethink:"sex"`
	BirthDate string    `json:"birth_date" gorethink:"birth_date"`
	ArmId     string    `json:"arm_id" gorethink:"arm_id"`
	FirstDate string    `json:"first_date" gorethink:"first_date"`
	IsDelete  bool      `json:"is_delete" gorethink:"is_delete"`
	TCreate   time.Time `json:"t_create" gorethink:"t_create"`
	ActorId   string    `json:"actor_id" gorethink:"actor_id"`
	StudyId   string    `json:"study_id" gorethink:"study_id"`
}

type SubjectTable struct {
	DBTable
}

func NewSubjectTable(session *gorethink.Session, db gorethink.Term) *SubjectTable {
	st := &SubjectTable{
		DBTable: DBTable{
			session: session,
			db:      db,
		},
	}
	st.Init("subject", "study_id", "is_delete")
	return st
}

func (st *SubjectTable) Insert(
	StudyId string,
	Name string,
	ScrNo string,
	Sex string,
	BirthDate string,
	ArmId string,
	FirstDate string,
	TCreate time.Time,
	ActorId string,
) (*Subject, error) {
	subject := &Subject{
		Name:      Name,
		ScrNo:     ScrNo,
		Sex:       Sex,
		BirthDate: BirthDate,
		ArmId:     ArmId,
		FirstDate: FirstDate,
		IsDelete:  false,
		TCreate:   TCreate,
		ActorId:   ActorId,
		StudyId:   StudyId,
	}

	_, err := st.SubjectByScrNo(ScrNo)
	if err != nil {
		if err != ErrNotExist {
			return nil, err
		}
	} else {
		return nil, ErrExistSubject
	}

	id, err := st.DBTable.Insert(subject)
	if err != nil {
		return nil, err
	}
	subject.Id = id
	return subject, nil
}

func (st *SubjectTable) Subject(SubjectId string) (*Subject, error) {
	res, err := st.table.
		Get(SubjectId).
		Run(st.session)
	if err != nil {
		return nil, err
	}

	var subject *Subject
	res.One(&subject)
	if len(subject.Id) == 0 {
		return nil, ErrNotExist
	}
	return subject, nil
}

func (st *SubjectTable) SubjectByScrNo(ScrNo string) (*Subject, error) {
	res, err := st.table.
		Filter(map[string]string{"scrno": ScrNo}).
		Run(st.Session())
	if err != nil {
		return nil, err
	}

	var subject *Subject
	err = res.One(&subject)
	if err != nil {
		if err == gorethink.ErrEmptyResult {
			return nil, ErrNotExist
		} else {
			return nil, err
		}
	}
	return subject, nil
}

func (st *SubjectTable) List(StudyId string) ([]*Subject, error) {
	res, err := st.table.
		GetAllByIndex("study_id", StudyId).
		Run(st.session)
	if err != nil {
		return nil, err
	}

	list := make([]*Subject, 0)
	res.All(&list)

	return list, nil
}

type Stack struct {
	Id        string    `json:"id,omitempty" gorethink:"id,omitempty"`
	FormId    string    `json:"form_id" gorethink:"form_id"`
	TCreate   time.Time `json:"t_create" gorethink:"t_create"`
	ActorId   string    `json:"actor_id" gorethink:"actor_id"`
	SubjectId string    `json:"subject_id" gorethink:"subject_id"`
}

type StackTable struct {
	DBTable
}

func NewStackTable(session *gorethink.Session, db gorethink.Term) *StackTable {
	st := &StackTable{
		DBTable: DBTable{
			session: session,
			db:      db,
		},
	}
	st.Init("stack", "subject_id")
	return st
}

func (st *StackTable) Stack(SubjectId string, FormId string) (*Stack, error) {
	res, err := st.table.
		GetAllByIndex("subject_id", SubjectId).
		Filter(map[string]interface{}{"form_id": FormId}).
		Run(st.session)
	if err != nil {
		return nil, err
	}

	var stack *Stack
	res.One(&stack)

	return stack, nil
}

func (st *StackTable) Insert(SubjectId string, FormId string, TCreate time.Time, ActorId string) (*Stack, error) {
	stack := &Stack{
		FormId:    FormId,
		TCreate:   TCreate,
		ActorId:   ActorId,
		SubjectId: SubjectId,
	}
	id, err := st.DBTable.Insert(stack)
	if err != nil {
		return nil, err
	}
	stack.Id = id
	return stack, nil
}

func (st *StackTable) List(SubjectId string) ([]*Stack, error) {
	res, err := st.table.
		GetAllByIndex("subject_id", SubjectId).
		Run(st.session)
	if err != nil {
		return nil, err
	}

	list := make([]*Stack, 0)
	res.All(&list)

	return list, nil
}

type Visit struct {
	Id       string    `json:"id,omitempty" gorethink:"id,omitempty"`
	Position string    `json:"position" gorethink:"position"`
	TCreate  time.Time `json:"t_create" gorethink:"t_create"`
	ActorId  string    `json:"actor_id" gorethink:"actor_id"`
	StackId  string    `json:"stack_id" gorethink:"stack_id"`
}

type VisitTable struct {
	DBTable
}

func NewVisitTable(session *gorethink.Session, db gorethink.Term) *VisitTable {
	st := &VisitTable{
		DBTable: DBTable{
			session: session,
			db:      db,
		},
	}
	st.Init("visit", "stack_id")
	return st
}

func (st *VisitTable) Insert(StackId string, Position string, TCreate time.Time, ActorId string) (*Visit, error) {
	visit := &Visit{
		Position: Position,
		TCreate:  TCreate,
		ActorId:  ActorId,
		StackId:  StackId,
	}
	id, err := st.DBTable.Insert(visit)
	if err != nil {
		return nil, err
	}
	visit.Id = id
	return visit, nil
}

func (st *VisitTable) List(StackId string) ([]*Visit, error) {
	res, err := st.table.
		GetAllByIndex("stack_id", StackId).
		Run(st.session)
	if err != nil {
		return nil, err
	}

	list := make([]*Visit, 0)
	res.All(&list)

	return list, nil
}

func (st *VisitTable) Visit(StackId string, Position string) (*Visit, error) {
	res, err := st.table.
		GetAllByIndex("stack_id", StackId).
		Filter(map[string]interface{}{"position": Position}).
		Run(st.session)
	if err != nil {
		return nil, err
	}

	var visit *Visit
	res.One(&visit)

	return visit, nil
}

type Data struct {
	Id       string    `json:"id,omitempty" gorethink:"id,omitempty"`
	Value    string    `json:"value,omitempty" gorethink:"value,omitempty"`
	CodeId   string    `json:"codeid,omitempty" gorethink:"codeid,omitempty"`
	Rowindex int64     `json:"rowindex" gorethink:"rowindex"`
	ItemId   string    `json:"itemid" gorethink:"itemid"`
	TCreate  time.Time `json:"t_create" gorethink:"t_create"`
	ActorId  string    `json:"actor_id" gorethink:"actor_id"`
	VisitId  string    `json:"visit_id" gorethink:"visit_id"`
}

func (data *Data) Clone() *Data {
	return &Data{
		Id:       data.Id,
		Value:    data.Value,
		CodeId:   data.CodeId,
		Rowindex: data.Rowindex,
		ItemId:   data.ItemId,
		TCreate:  data.TCreate,
		ActorId:  data.ActorId,
		VisitId:  data.VisitId,
	}
}

type DataTable struct {
	DBTable
}

func NewDataTable(session *gorethink.Session, db gorethink.Term) *DataTable {
	st := &DataTable{
		DBTable: DBTable{
			session: session,
			db:      db,
		},
	}
	st.Init("data", "visit_id")
	return st
}

func (st *DataTable) Insert(datas []*Data) error {
	err := st.InsertBatch(datas)
	if err != nil {
		return err
	}
	return nil
}

func (st *DataTable) List(VisitId string) ([]*Data, error) {
	res, err := st.table.
		GetAllByIndex("visit_id", VisitId).
		Run(st.session)
	if err != nil {
		return nil, err
	}

	list := make([]*Data, 0)
	res.All(&list)

	return list, nil
}

func (st *DataTable) ListByRowindex(VisitId string, Rowindex int64) ([]*Data, error) {
	res, err := st.table.
		GetAllByIndex("visit_id", VisitId).
		Filter(map[string]interface{}{"rowindex": Rowindex}).
		Run(st.session)
	if err != nil {
		return nil, err
	}

	list := make([]*Data, 0)
	res.All(&list)

	return list, nil
}

type History struct {
	Id      string    `json:"id,omitempty" gorethink:"id,omitempty"`
	Data    Data      `json:"data"`
	TCreate time.Time `json:"t_create" gorethink:"t_create"`
	ActorId string    `json:"actor_id" gorethink:"actor_id"`
	DataId  string    `json:"data_id" gorethink:"data_id"`
}

type HistoryTable struct {
	DBTable
}

func NewHistoryTable(session *gorethink.Session, db gorethink.Term) *HistoryTable {
	st := &HistoryTable{
		DBTable: DBTable{
			session: session,
			db:      db,
		},
	}
	st.Init("history", "data_id")
	return st
}

func (st *HistoryTable) Retain(data *Data, TCreate time.Time, ActorId string) *History {
	history := &History{
		Data:    (*data.Clone()),
		TCreate: TCreate,
		ActorId: ActorId,
	}
	return history
}

func (st *HistoryTable) Insert(histories []*History) error {
	err := st.InsertBatch(histories)
	if err != nil {
		return err
	}
	return nil
}

func (st *HistoryTable) List(VisitId string) ([]*History, error) {
	res, err := st.table.
		GetAllByIndex("visit_id", VisitId).
		Run(st.session)
	if err != nil {
		return nil, err
	}

	list := make([]*History, 0)
	res.All(&list)

	return list, nil
}

type Form struct {
	Id         string                 `json:"id,omitempty" gorethink:"id,omitempty"`
	Name       string                 `json:"name" gorethink:"name"`
	Type       string                 `json:"type" gorethink:"type"`
	Plan       *Plan                  `json:"plan" gorethink:"plan,omitempty""`
	Extra      map[string]string      `json:"extra" gorethink:"extra"`
	extraCache map[string]interface{} `json:"-" gorethink:"-"`
	Groups     []*Group               `json:"-" gorethink:"-"`
	StudyId    string                 `json:"study_id" gorethink:"study_id"`
}

type FormTable struct {
	DBTable
	gt *GroupTable
}

func NewFormTable(session *gorethink.Session, db gorethink.Term) *FormTable {
	st := &FormTable{
		DBTable: DBTable{
			session: session,
			db:      db,
		},
		gt: NewGroupTable(session, db),
	}

	st.Init("form", "study_id")
	return st
}

func (st *FormTable) Form(FormId string) (*Form, error) {
	res, err := st.table.
		Get(FormId).
		Run(st.session)
	if err != nil {
		return nil, err
	}

	if res.IsNil() {
		return nil, fmt.Errorf("not found")
	} else {
		form := &Form{}
		res.One(&form)

		st.FillGroup(form)

		return form, nil
	}
}

func (st *FormTable) List(StudyId string) ([]*Form, error) {
	res, err := st.table.
		GetAllByIndex("study_id", StudyId).
		OrderBy("priority").
		Run(st.session)
	if err != nil {
		return nil, err
	}

	list := make([]*Form, 0)
	res.All(&list)

	for _, form := range list {
		st.FillGroup(form)
	}

	return list, nil
}

func (st *FormTable) FillGroup(form *Form) error {
	list, err := st.gt.List(form.Id)
	if err != nil {
		return err
	}

	form.Groups = list
	return nil
}

func (f *Form) GenerateHTML(position string, n *htmlwriter.HtmlNode) error {
	if guide, has := f.Extra["guide"]; has {
		n.Add("div").Class("guide fold").Add("div").Class("guide-wrapper").Add("div").Class("guide-content").SetText(strings.Join(strings.Split(guide, "\n"), "<br/>"))
	}

	for i, g := range f.Groups {
		isEnable, err := g.IsEnable(position)
		if err != nil {
			return err
		}
		if isEnable {
			jGroup := n.Add("section").Class("form")
			isNextHidden := false
			if i+1 < len(f.Groups) {
				label := f.Groups[i+1].Extra["label"]
				if label == "hidden" {
					isNextHidden = true
				}
			}
			if isNextHidden {
				jGroup.Style("margin-bottom", "0px")
			}
			err := g.GenerateHTML(position, jGroup, false)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

/*
[C] : 조건이 (참/거짓)인 동안 / 조건이 (참/거짓)일 때까지 / 한 번이라도 조건이 (참/거짓)인 이후로 부터 / 한 번이라도 조건이 (참/거짓) 일 때까지
[D] : (최초/최종/특정방문) 일자

[C] 일 때 [D]로부터 0[T] 후 0-0[T] 동안 0[T] 마다
[C] 일 때 [D]로부터 0[T] 후 [A, B, C, D][T] 때
*/
type Plan struct {
	Condition    string       `json:"condition,omitempty" gorethink:"condition,omitempty"`
	DateItemId   string       `json:"date_item_id,omitempty" gorethink:"date_item_id,omitempty"`
	DateValue    string       `json:"date_value,omitempty" gorethink:"date_value,omitempty"`
	DatePosition string       `json:"date_position,omitempty" gorethink:"date_position,omitempty"`
	Delay        *PlanValue   `json:"delay,omitempty" gorethink:"delay,omitempty"`
	ValueFrom    *PlanValue   `json:"value_from,omitempty" gorethink:"value_form,omitempty"`
	ValueTo      *PlanValue   `json:"value_to,omitempty" gorethink:"value_to,omitempty"`
	Interval     *PlanValue   `json:"interval,omitempty" gorethink:"interval,omitempty"`
	ValueAt      []*PlanValue `json:"value_at,omitempty" gorethink:"value_at,omitempty"`
}

type PlanValue struct {
	Value string   `json:"value,omitempty" gorethink:"value,omitempty"`
	Unit  PlanUnit `json:"unit,omitempty" gorethink:"unit,omitempty"`
}

type PlanUnit string

const (
	Quantity PlanUnit = "quantity"
	Year     PlanUnit = "year"
	Month    PlanUnit = "month"
	Week     PlanUnit = "week"
	Weekday  PlanUnit = "weekday"
	Day      PlanUnit = "day"
	Hour     PlanUnit = "hour"
	Minute   PlanUnit = "minute"
	Second   PlanUnit = "second"
)

type Group struct {
	Id         string                 `json:"id,omitempty" gorethink:"id,omitempty"`
	Name       string                 `json:"name" gorethink:"name"`
	Type       string                 `json:"type" gorethink:"type"`
	Extra      map[string]string      `json:"extra" gorethink:"extra"`
	extraCache map[string]interface{} `json:"-" gorethink:"-"`
	Items      []*Item                `json:"-" gorethink:"-"`
	FormId     string                 `json:"form_id" gorethink:"form_id"`
}

type GroupTable struct {
	DBTable
	it *ItemTable
}

func NewGroupTable(session *gorethink.Session, db gorethink.Term) *GroupTable {
	st := &GroupTable{
		DBTable: DBTable{
			session: session,
			db:      db,
		},
		it: NewItemTable(session, db),
	}
	st.Init("group", "form_id")
	return st
}

func (st *GroupTable) List(FormId string) ([]*Group, error) {
	res, err := st.table.
		GetAllByIndex("form_id", FormId).
		OrderBy("priority").
		Run(st.session)
	if err != nil {
		return nil, err
	}

	list := make([]*Group, 0)
	res.All(&list)

	for _, group := range list {
		st.FillItem(group)
	}

	return list, nil
}

func (gt *GroupTable) FillItem(group *Group) error {
	list, err := gt.it.List(group.Id)
	if err != nil {
		return err
	}

	group.Items = list
	return nil
}

func (g *Group) VisitName(position string) (string, error) {
	if g.extraCache == nil {
		g.extraCache = make(map[string]interface{})
	}
	return visitNameByExtra(g.Extra, g.extraCache, position, g.Name, g)
}
func (g *Group) IsEnable(position string) (bool, error) {
	if g.extraCache == nil {
		g.extraCache = make(map[string]interface{})
	}
	isEnable, err := checkEnableByExtra(g.Extra, g.extraCache, position, g)
	if err != nil {
		return false, err
	}

	if isEnable {
		hasEnable := false
		for _, item := range g.Items {
			isEnable, err := item.IsEnable(position)
			if err != nil {
				return false, err
			}
			if isEnable {
				hasEnable = true
				break
			}
		}
		isEnable = hasEnable
	}
	return isEnable, nil
}

func (g *Group) GenerateHTML(position string, jGroup *htmlwriter.HtmlNode, isListPage bool) error {
	jGroup.Attr("groupid", g.Id)
	jGroup.Class(fmt.Sprintf("group-%s", g.Type))

	if hidden := g.Extra["hidden"]; hidden == "1" {
		jGroup.Style("display", "none")
	}

	label := g.Extra["label"]

	showName := (label != "hidden" && g.Type != "list")
	if showName {
		jHead := jGroup.Add("div").Class("content-block-title")

		name, err := g.VisitName(position)
		if err != nil {
			return err
		}
		jHead.SetText(name)

		tooltip := g.Extra["tooltip"]
		if tooltipVisit, has := g.Extra["tooltip.visit"]; has {
			var hash map[string]string
			err := json.Unmarshal([]byte(tooltipVisit), &hash)
			if err != nil {
				log.Println("json error", g, tooltipVisit)
				return err
			}
			if v, has := hash[position]; has {
				tooltip = v
			}
		}
		if len(tooltip) > 0 {
			jTooltip := jGroup.Add("div").Class("content-block-inner").SetText(tooltip)
			if tooltipColor, has := g.Extra["tooltip.color"]; has {
				jTooltip.Style("color", tooltipColor)
			}
		}
	}

	switch g.Type {
	case "list":
		return g.iterListChildren(position, jGroup, isListPage)
	default:
		return g.iterChildren(position, jGroup)
	}
}
func (g *Group) iterChildren(position string, jGroup *htmlwriter.HtmlNode) error {
	jUl := jGroup.Add("div").Class("list-block form-type-normal").Style("margin-top", "0px").Add("ul")
	for _, item := range g.Items {
		isEnable, err := item.IsEnable(position)
		if err != nil {
			return err
		}
		if isEnable {
			jLi := jUl.Add("li").Class("item-block")
			var jName *htmlwriter.HtmlNode
			showName := true
			if label := item.Extra["label"]; label == "hidden" {
				showName = false
			}
			if showName {
				jName = jLi.Add("div").Class("content-block-inner")
			}
			jDiv := jLi.Add("div").Class("item-content").Style("padding", "0px")
			jItem := jDiv.Add("div").Class("item-inner").Style("padding", "0px")

			err := item.GenerateHTML(position, jName, jItem)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
func (g *Group) iterListChildren(position string, jGroup *htmlwriter.HtmlNode, isListPage bool) error {
	jUl := jGroup.Add("div").Class("list-block form-type-normal").Style("margin-top", "0px").Add("ul")
	for _, item := range g.Items {
		isEnable, err := item.IsEnable(position)
		if err != nil {
			return err
		}
		if isEnable {
			jLi := jUl.Add("li").Class("item-block")
			var jName *htmlwriter.HtmlNode
			showName := !isListPage
			if label := item.Extra["label"]; label == "hidden" {
				showName = false
			}
			if showName {
				jName = jLi.Add("div").Class("content-block-inner")
			}
			jDiv := jLi.Add("div").Class("item-content").Style("padding", "0px")
			jItem := jDiv.Add("div").Class("item-inner").Style("padding", "0px")

			err := item.GenerateHTML(position, jName, jItem)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

type Item struct {
	Id         string                 `json:"id,omitempty" gorethink:"id,omitempty"`
	Name       string                 `json:"name" gorethink:"name"`
	Type       string                 `json:"type" gorethink:"type"`
	TypeArgv   string                 `json:"type_argv" gorethink:"type_argv"`
	Extra      map[string]string      `json:"extra" gorethink:"extra"`
	extraCache map[string]interface{} `json:"-" gorethink:"-"`
	Codes      []*Code                `json:"-" gorethink:"-"`
	GroupId    string                 `json:"group_id" gorethink:"group_id"`
}

type ItemTable struct {
	DBTable
	ct *CodeTable
}

func NewItemTable(session *gorethink.Session, db gorethink.Term) *ItemTable {
	st := &ItemTable{
		DBTable: DBTable{
			session: session,
			db:      db,
		},
		ct: NewCodeTable(session, db),
	}
	st.Init("item", "group_id")
	return st
}

func (st *ItemTable) List(GroupId string) ([]*Item, error) {
	res, err := st.table.
		GetAllByIndex("group_id", GroupId).
		OrderBy("priority").
		Run(st.session)
	if err != nil {
		return nil, err
	}

	list := make([]*Item, 0)
	res.All(&list)

	for _, item := range list {
		st.FillCode(item)
	}

	return list, nil
}

func (it *ItemTable) FillCode(item *Item) error {
	list, err := it.ct.List(item.Id)
	if err != nil {
		return err
	}

	item.Codes = list
	return nil
}

func (item *Item) VisitName(position string) (string, error) {
	if item.extraCache == nil {
		item.extraCache = make(map[string]interface{})
	}
	return visitNameByExtra(item.Extra, item.extraCache, position, item.Name, item)
}
func (item *Item) IsEnable(position string) (bool, error) {
	if item.extraCache == nil {
		item.extraCache = make(map[string]interface{})
	}
	isEnable, err := checkEnableByExtra(item.Extra, item.extraCache, position, item)
	if err != nil {
		return false, err
	}
	if isEnable && len(item.Codes) > 0 {
		hasEnable := false
		for _, code := range item.Codes {
			isEnable, err := code.IsEnable(position)
			if err != nil {
				return false, err
			}
			if isEnable {
				hasEnable = true
				break
			}
		}
		isEnable = hasEnable
	}
	return isEnable, nil
}

func (item *Item) GenerateHTML(position string, jName *htmlwriter.HtmlNode, jItem *htmlwriter.HtmlNode) error {
	dataKey := item.Id
	jItem.Class(fmt.Sprintf("item-%s", item.Id))

	showName := (jName != nil && item.Type != "label")
	if showName {
		jTable := jName.Add("table")
		jTr := jTable.Add("tr")
		if category, has := item.Extra["category"]; has {
			jCategory := jTr.Add("th").Class("item-category").Style("text-align", "center")
			for i, v := range strings.Split(category, "/") {
				if i > 0 {
					jCategory.Add("hr").Attr("noshade", "")
				}
				jCategory.Add("lable").SetText(v)
			}
		}

		name, err := item.VisitName(position)
		if err != nil {
			return err
		}
		jTr.Add("td").SetText(name)

		if width, has := item.Extra["width"]; has {
			jName.Style("width", width)
		}
	}

	if require := item.Extra["require"]; require == "1" {
		if showName {
			jName.Class("required")
		} else {
			jItem.Class("required")
		}
		jItem.Class("require-item")
	}

	switch item.Type {
	case "label":
		jItem.Add("span").SetText(item.Name)
	case "checkbox":
		fallthrough
	case "radio":
		//handle option
		jUl := jItem.Add("ul").Style("padding-left", "0px").Style("width", "100%")
		for _, code := range item.Codes {
			isEnable, err := code.IsEnable(position)
			if err != nil {
				return err
			}
			if isEnable {
				jCode := jUl.Add("li").Add("label").Class("label-radio item-content")
				jInput := jCode.Add("input").Attr("type", item.Type).Style("cursor", "pointer")
				jInput.Attr("name", dataKey)
				jInput.Attr("value", code.Id)

				name, err := code.VisitName(position)
				if err != nil {
					return err
				}

				jDiv := jCode.Add("div").Class("item-inner").Style("margin-left", "0px").Style("padding", "0px")
				jLabel := jDiv.Add("div").Class("item-title").Class("item-title")
				jLabel.Style("width", "100%").Style("height", "100%").Style("padding", "6px").Style("line-height", "20px").Style("white-space", "normal")
				jLabel.Style("text-align", "left").SetText(name)
				if follow, has := code.Extra["follow"]; has {
					jInput.Attr("follow", follow)
				}
			}
		}
	case "textarea":
		jTextarea := jItem.Add("textarea").Attr("type", "text").Attr("itemtype", item.Type).Class("form-item")
		jTextarea.Attr("name", dataKey)
		if tooltip, has := item.Extra["tooltip"]; has {
			jTextarea.Attr("placeholder", tooltip)
		}
	default:
		jInput := jItem.Add("input").Attr("type", "text").Attr("itemtype", item.Type).Class("form-item")
		jInput.Attr("name", dataKey)
		if tooltip, has := item.Extra["tooltip"]; has {
			jInput.Attr("placeholder", tooltip)
		}

		switch item.Type {
		case "date":
			jInput.Attr("type", "date")
		case "script":
			jInput.Attr("readonly", "")
		}

		if float, has := item.Extra["float"]; has {
			jInput.Attr("float", float)
		}
		if unit, has := item.Extra["unit"]; has {
			jItem.Add("label").SetText(unit)
		}

		if formula, has := item.Extra["formula"]; has {
			jInput.Attr("formula", formula)
			jInput.Attr("readonly", "readonly")
		}

		if r, has := item.Extra["range"]; has {
			jInput.Attr("range", r)
		}
	}

	if readonly := item.Extra["readonly"]; readonly == "1" {
		jItem.Attr("readonly", "")
	}
	return nil
}

type Code struct {
	Id         string                 `json:"id,omitempty" gorethink:"id,omitempty"`
	Name       string                 `json:"name" gorethink:"name"`
	Value      string                 `json:"value" gorethink:"value"`
	Extra      map[string]string      `json:"extra" gorethink:"extra"`
	extraCache map[string]interface{} `json:"-" gorethink:"-"`
	ItemId     string                 `json:"item_id" gorethink:"item_id"`
}

type CodeTable struct {
	DBTable
}

func NewCodeTable(session *gorethink.Session, db gorethink.Term) *CodeTable {
	st := &CodeTable{
		DBTable: DBTable{
			session: session,
			db:      db,
		},
	}
	st.Init("code", "item_id")
	return st
}

func (st *CodeTable) List(ItemId string) ([]*Code, error) {
	res, err := st.table.
		GetAllByIndex("item_id", ItemId).
		OrderBy("priority").
		Run(st.session)
	if err != nil {
		return nil, err
	}

	list := make([]*Code, 0)
	res.All(&list)

	return list, nil
}

func (code *Code) VisitName(position string) (string, error) {
	if code.extraCache == nil {
		code.extraCache = make(map[string]interface{})
	}
	return visitNameByExtra(code.Extra, code.extraCache, position, code.Name, code)
}
func (code *Code) IsEnable(position string) (bool, error) {
	if code.extraCache == nil {
		code.extraCache = make(map[string]interface{})
	}
	return checkEnableByExtra(code.Extra, code.extraCache, position, code)
}

func visitNameByExtra(Extra map[string]string, extraCache map[string]interface{}, position string, name string, debug interface{}) (string, error) {
	key := "name@" + position
	cache, has := extraCache[key]
	if has {
		return cache.(string), nil
	} else {
		visitName := name
		if nameVisit, has := Extra["name.visit"]; has {
			var hash map[string]string
			err := json.Unmarshal([]byte(nameVisit), &hash)
			if err != nil {
				log.Println("json error", debug, nameVisit)
				return "", err
			}
			if v, has := hash[position]; has {
				visitName = v
			}
		}
		extraCache[key] = visitName
		return visitName, nil
	}
}
func checkEnableByExtra(Extra map[string]string, extraCache map[string]interface{}, position string, debug interface{}) (bool, error) {
	key := "enable@" + position
	cache, has := extraCache[key]
	if has {
		return cache.(bool), nil
	} else {
		isAllowMode := true
		var visitData string
		if visitAllow, has := Extra["visit.allow"]; has {
			visitData = visitAllow
		} else if visitDeny, has := Extra["visit.deny"]; has {
			isAllowMode = false
			visitData = visitDeny
		}

		isEnable := true
		if len(visitData) > 0 {
			var list []interface{}
			err := json.Unmarshal([]byte(visitData), &list)
			if err != nil {
				log.Println("json error", debug, visitData)
				return false, err
			}

			if isAllowMode {
				isEnable = false
			}
			for _, v := range list {
				if position == convert.String(v) {
					isEnable = isAllowMode
				}
			}
		}
		extraCache[key] = isEnable
		return isEnable, nil
	}
}
