package main

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"strings"
	"time"

	"git.tnicdev.co.kr/hami/CTC_Web/pkg/util"

	"github.com/PuerkitoBio/goquery"
	"github.com/blackss2/utility/convert"
	"github.com/blackss2/utility/htmlwriter"
	"github.com/labstack/echo"
)

func route_api_subjects(g *echo.Group) {
	g.GET("/subjects", func(c echo.Context) error {
		retCode, retValue := (func() (int, interface{}) {
			scrno, err := util.ParamToString(c, "scrno", false)
			if err != nil {
				return http.StatusBadRequest, err
			}
			arm, err := util.ParamToString(c, "arm", false)
			if err != nil {
				return http.StatusBadRequest, err
			}
			Type, err := util.ParamToString(c, "type", false)
			if err != nil {
				return http.StatusBadRequest, err
			}
			//////////////////////////////////////////////////

			subjects, err := Search_Subjects(scrno, arm, Type)
			if err != nil {
				return http.StatusInternalServerError, err
			}
			return http.StatusOK, subjects
		})()

		if err, is := retValue.(error); is {
			return c.JSON(retCode, &Result{Error: err})
		} else {
			return c.JSON(retCode, &Result{Result: retValue})
		}
	})
	g.GET("/subjects/:subjectid", func(c echo.Context) error {
		retCode, retValue := (func() (int, interface{}) {
			subjectid, err := util.PathToString(c, "subjectid")
			if err != nil {
				return http.StatusBadRequest, err
			}
			//////////////////////////////////////////////////

			subject, err := gAPI.SubjectTable.Subject(subjectid)
			if err != nil {
				return http.StatusInternalServerError, err
			}
			if subject.StudyId != gConfig.StudyId {
				return http.StatusBadRequest, ErrNotMatchedStudyId
			}

			dao, err := Select_Subject(subject)
			if err != nil {
				return http.StatusInternalServerError, err
			}
			return http.StatusOK, dao
		})()

		if err, is := retValue.(error); is {
			return c.JSON(retCode, &Result{Error: err})
		} else {
			return c.JSON(retCode, &Result{Result: retValue})
		}
	})
	g.GET("/subjects/:subjectid/forms/:formid/groups/:groupid", func(c echo.Context) error {
		Uid := c.Get(UID_KEY).(string)

		retCode, retValue := (func() (int, interface{}) {
			subjectid, err := util.PathToString(c, "subjectid")
			if err != nil {
				return http.StatusBadRequest, err
			}
			formid, err := util.PathToString(c, "formid")
			if err != nil {
				return http.StatusBadRequest, err
			}
			groupid, err := util.PathToString(c, "groupid")
			if err != nil {
				return http.StatusBadRequest, err
			}
			position, err := util.ParamToString(c, "position", true)
			if err != nil {
				return http.StatusBadRequest, err
			}
			//////////////////////////////////////////////////

			subject, err := gAPI.SubjectTable.Subject(subjectid)
			if err != nil {
				return http.StatusInternalServerError, err
			}
			if subject.StudyId != gConfig.StudyId {
				return http.StatusBadRequest, ErrNotMatchedStudyId
			}

			forms, err := FormWithCache(gAPI.FormTable, gConfig.StudyId)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			var form *Form
			for _, f := range forms {
				if f.Id == formid {
					form = f
					break
				}
			}
			if form == nil {
				return http.StatusInternalServerError, ErrNotExistForm
			}

			TNow := time.Now()

			stack, err := gAPI.StackTable.Stack(subject.Id, form.Id)
			if err != nil {
				if err != ErrNotExist {
					return http.StatusInternalServerError, err
				}
			}
			if stack == nil {
				s, err := gAPI.StackTable.Insert(subject.Id, form.Id, TNow, Uid)
				if err != nil {
					return http.StatusInternalServerError, err
				}
				stack = s
			}

			visit, err := gAPI.VisitTable.Visit(stack.Id, position)
			if err != nil {
				if err != ErrNotExist {
					return http.StatusInternalServerError, err
				}
			}
			if visit == nil {
				v, err := gAPI.VisitTable.Insert(stack.Id, position, TNow, Uid)
				if err != nil {
					return http.StatusInternalServerError, err
				}
				visit = v
			}

			dataList, err := gAPI.DataTable.List(visit.Id)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			groupHash := make(map[string]*Group)
			itemHash := make(map[string]*Item)
			for _, form := range forms {
				addFormMeta(form, groupHash, itemHash)
			}

			var group *Group
			for _, g := range form.Groups {
				if g.Id == groupid {
					group = g
					break
				}
			}

			if group == nil {
				return http.StatusInternalServerError, ErrNotExistGroup
			}

			var dataRowindexList [][]*Data
			if len(dataList) > 0 {
				maxRowindex := int64(0)
				dataHash := make(map[int64][]*Data)
				for _, data := range dataList {
					if item, has := itemHash[data.ItemId]; has {
						if item.GroupId == group.Id {
							if maxRowindex < data.Rowindex {
								maxRowindex = data.Rowindex
							}
							list, has := dataHash[data.Rowindex]
							if !has {
								list = make([]*Data, 0)
							}
							dataHash[data.Rowindex] = append(list, data)
						}
					}
				}
				dataRowindexList = make([][]*Data, maxRowindex+1)
				for rowindex, list := range dataHash {
					dataRowindexList[rowindex] = list
				}
			} /* else if group.Type != "list" {
				dataRowindexList = [][]*Data{[]*Data{}}
			}
			*/

			jRoot := htmlwriter.CreateHtmlNode("div").Class("form-grp")
			jRoot.Attr("formid", form.Id)
			jRoot.Attr("position", position)
			err = group.GenerateHTML(position, jRoot, false)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			var buffer bytes.Buffer
			jRoot.WriteWith(&buffer, "\t")

			var htmlBuffer bytes.Buffer
			if true {
				docOrg, err := goquery.NewDocumentFromReader(&buffer)
				if err != nil {
					return http.StatusInternalServerError, err
				}

				for r, list := range dataRowindexList {
					if list == nil {
						continue
					}

					//apply data to html
					doc := docOrg.Clone()
					jFormGrp := doc.Find(".form-grp")

					jFormGrp.SetAttr("rowindex", convert.String(r))
					for _, data := range list {
						if len(data.Value) > 0 || len(data.CodeId) > 0 {
							item, has := itemHash[data.ItemId]
							if !has {
								return http.StatusInternalServerError, fmt.Errorf("no item : %v", data.ItemId)
							}

							formKey := data.ItemId
							jTarget := jFormGrp.Find(fmt.Sprintf("[name='%s']", formKey))
							if jTarget.Length() == 0 {
								return http.StatusInternalServerError, fmt.Errorf("target length is zero")
							}

							var Value string
							switch item.Type {
							case "checkbox":
								fallthrough
							case "radio":
								Value = data.CodeId
							default:
								Value = data.Value
							}

							switch item.Type {
							case "checkbox":
								fallthrough
							case "radio":
								jTarget = jTarget.Filter(fmt.Sprintf("[value='%s']", Value))
								jTarget.SetAttr("checked", "checked")
							case "textarea":
								jTarget.AppendHtml(Value)
							default:
								jTarget.SetAttr("value", Value)
							}
						}
					}

					ret, err := doc.Html()
					if err != nil {
						return http.StatusInternalServerError, err
					}

					htmlBuffer.WriteString(ret)
				}
			}

			return http.StatusOK, template.HTML(htmlBuffer.String())
		})()

		if err, is := retValue.(error); is {
			return c.JSON(retCode, &Result{Error: err})
		} else {
			return c.JSON(retCode, &Result{Result: retValue})
		}
	})
	g.POST("/subjects", func(c echo.Context) error {
		Uid := c.Get(UID_KEY).(string)

		retCode, retValue := (func() (int, interface{}) {
			var item struct {
				Name      string `json:"name"`
				ScrNo     string `json:"scrno"`
				Sex       string `json:"sex"`
				BirthDate string `json:"birth_date"`
				Arm       string `json:"arm"`
				FirstDate string `json:"first_date"`
			}
			err := util.BodyToStruct(c.Request().Body, &item)
			if err != nil {
				return http.StatusBadRequest, err
			}
			//////////////////////////////////////////////////

			_, err = gAPI.SubjectTable.Insert(gConfig.StudyId, item.Name, item.ScrNo, item.Sex, item.BirthDate, item.Arm, item.FirstDate, time.Now(), Uid)
			if err != nil {
				return http.StatusInternalServerError, err
			}
			return http.StatusOK, true
		})()
		if err, is := retValue.(error); is {
			return c.JSON(retCode, &Result{Error: err})
		} else {
			return c.JSON(retCode, &Result{Result: retValue})
		}
	})
	g.PUT("/subjects/:subjectid", func(c echo.Context) error {
		retCode, retValue := (func() (int, interface{}) {
			subjectid, err := util.PathToString(c, "subjectid")
			if err != nil {
				return http.StatusBadRequest, err
			}

			var item struct {
				Name      string `json:"name"`
				ScrNo     string `json:"scrno"`
				Sex       string `json:"sex"`
				BirthDate string `json:"birth_date"`
				FirstDate string `json:"first_date"`
			}
			err = util.BodyToStruct(c.Request().Body, &item)
			if err != nil {
				return http.StatusBadRequest, err
			}
			//////////////////////////////////////////////////

			IsAuth := false //TEMP

			subject, err := gAPI.SubjectTable.Subject(subjectid)
			if err != nil {
				return http.StatusInternalServerError, err
			}
			if subject.StudyId != gConfig.StudyId {
				return http.StatusBadRequest, ErrNotMatchedStudyId
			}

			if IsAuth {
				subject.Name = item.Name
			}
			subject.ScrNo = item.ScrNo
			subject.Sex = item.Sex
			subject.BirthDate = item.BirthDate
			subject.FirstDate = item.FirstDate

			err = gAPI.SubjectTable.Update(subject.Id, subject)
			if err != nil {
				return http.StatusInternalServerError, err
			}
			dao, err := Select_Subject(subject)
			if err != nil {
				return http.StatusInternalServerError, err
			}
			return http.StatusOK, dao
		})()
		if err, is := retValue.(error); is {
			return c.JSON(retCode, &Result{Error: err})
		} else {
			return c.JSON(retCode, &Result{Result: retValue})
		}
	})
}

type DAO_Search_Subject struct {
	Id         string `json:"id,omitempty"`
	ScrNo      string `json:"scrno"`
	Age        string `json:"age"`
	Sex        string `json:"sex"`
	Tag        string `json:"tag"`
	Compliance int    `json:"compliance"`
	Progress   string `json:"progress"`
	TCreate    string `json:"t_create"`
}

func Search_Subjects(scrno string, arm string, Type string) ([]*DAO_Search_Subject, error) {
	list, err := gAPI.SubjectTable.List(gConfig.StudyId)
	if err != nil {
		return nil, err
	}

	//Type

	daos := make([]*DAO_Search_Subject, 0)
	for _, v := range list {
		if len(scrno) > 0 {
			if !strings.Contains(v.ScrNo, scrno) {
				continue
			}
		}
		if len(arm) > 0 {
			if !strings.Contains(v.Arm, arm) {
				continue
			}
		}

		hasAE := false
		if true {
			count, err := Subject_AECount(v.Id)
			if err != nil {
				return nil, err
			}

			if count > 0 {
				hasAE = true
			}
		}
		hasLow := false
		compliance := 0
		maxPosition := int64(0)
		if true {
			cps, err := Subject_Compliance(v.Id)
			if err != nil {
				return nil, err
			}

			totalCount := 0
			totalItemCount := 0
			for _, v := range cps {
				if v.HasVisit {
					totalCount += v.TotalCount
					totalItemCount += v.TotalItemCount

					p := convert.Int(v.Day)
					if maxPosition < p {
						maxPosition = p
					}
				}
			}

			if totalItemCount > 0 {
				LOW_CUTOFF := 70

				compliance = totalCount * 100 / totalItemCount
				if compliance < LOW_CUTOFF {
					hasLow = true
				}
			}
		}
		if len(Type) > 0 {
			switch Type {
			case "ae":
				if !hasAE {
					continue
				}
			case "low":
				if !hasLow {
					continue
				}
			}
		}

		tag := ""
		if hasAE {
			tag = "ae"
		} else if hasLow {
			tag = "low"
		}

		dao := &DAO_Search_Subject{
			Id:         v.Id,
			ScrNo:      v.ScrNo,
			Age:        getAgeFromDate(v.BirthDate),
			Sex:        v.Sex,
			Tag:        tag,
			Compliance: compliance,
			Progress:   fmt.Sprintf("%d일차 %d%%", maxPosition, compliance), //TEMP
			TCreate:    convert.String(v.TCreate)[:10],
		}
		daos = append(daos, dao)
	}
	sort.Sort(DAO_Search_Subject_Sort(daos))
	return daos, nil
}

type DAO_Search_Subject_Sort []*DAO_Search_Subject

func (s DAO_Search_Subject_Sort) Len() int {
	return len(s)
}

func (s DAO_Search_Subject_Sort) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s DAO_Search_Subject_Sort) Less(i, j int) bool {
	return s[i].ScrNo < s[j].ScrNo
}

type DAO_Select_Subject struct {
	Id          string         `json:"id,omitempty"`
	Name        string         `json:"name"`
	ScrNo       string         `json:"scrno"`
	Age         string         `json:"age"`
	Sex         string         `json:"sex"`
	Compliance  map[string]int `json:"compliance"`
	BirthDate   string         `json:"birth_date"`
	FirstDate   string         `json:"first_date"`
	MaxPosition int64          `json:"max_position"`
	IsAuth      bool           `json:"is_auth"`
	TCreate     string         `json:"t_create"`
}

func Select_Subject(subject *Subject) (*DAO_Select_Subject, error) {
	cps, err := Subject_Compliance(subject.Id)
	if err != nil {
		return nil, err
	}

	compliance := make(map[string]int)
	maxPosition := int64(0)
	for _, v := range cps {
		if v.TotalItemCount > 0 {
			compliance[v.Day] = v.TotalCount * 100 / v.TotalItemCount
		}
		if v.HasVisit {
			p := convert.Int(v.Day)
			if maxPosition < p {
				maxPosition = p
			}
		}
	}

	dao := &DAO_Select_Subject{
		Id:          subject.Id,
		ScrNo:       subject.ScrNo,
		Age:         getAgeFromDate(subject.BirthDate),
		Sex:         subject.Sex,
		Compliance:  compliance,
		BirthDate:   subject.BirthDate,
		FirstDate:   subject.FirstDate,
		MaxPosition: maxPosition,
		IsAuth:      false, //TEMP
		TCreate:     convert.String(subject.TCreate)[:10],
	}
	if dao.IsAuth {
		dao.Name = subject.Name
	}
	return dao, nil
}

func getAgeFromDate(date string) string {
	Age := ""
	t := convert.Time(date)
	if t != nil {
		now := time.Now()
		c := 0
		if int(now.Month())*100+int(now.Day()) >= int(t.Month())*100+int(t.Day()) {
			c++
		}
		Age = convert.String((now.Year() - t.Year()) + c)
	}
	return Age
}
