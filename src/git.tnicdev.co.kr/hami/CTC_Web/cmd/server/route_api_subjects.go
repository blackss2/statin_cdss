package main

import (
	"bytes"
	"net/http"
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
			name, err := util.ParamToString(c, "name", false)
			if err != nil {
				return http.StatusBadRequest, err
			}
			scrno, err := util.ParamToString(c, "scrno", false)
			if err != nil {
				return http.StatusBadRequest, err
			}
			arm_id, err := util.ParamToString(c, "arm_id", false)
			if err != nil {
				return http.StatusBadRequest, err
			}
			//////////////////////////////////////////////////

			subjects, err := Search_Subjects(name, scrno, arm_id)
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

			subject, err := Select_Subjects(subjectid)
			if err != nil {
				return http.StatusInternalServerError, err
			}
			return http.StatusOK, subject
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
			rowindex, err := util.ParamToInt(c, "rowindex", true)
			if err != nil {
				return http.StatusBadRequest, err
			}
			//////////////////////////////////////////////////

			subject, err := gAPI.SubjectTable.Subject(subjectid)
			if err != nil {
				return http.StatusInternalServerError, err
			}
			if subject.StudyId != gConfig.StudyId {
				return http.StatusInternalServerError, ErrNotMatchedStudyId
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
				return http.StatusInternalServerError, err
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
				return http.StatusInternalServerError, err
			}
			if visit == nil {
				v, err := gAPI.VisitTable.Insert(stack.Id, position, TNow, Uid)
				if err != nil {
					return http.StatusInternalServerError, err
				}
				visit = v
			}

			dataList, err := gAPI.DataTable.ListByRowindex(visit.Id, rowindex)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			groupHash := make(map[string]*Group)
			itemHash := make(map[string]*Item)
			for _, form := range forms {
				addFormMeta(form, groupHash, itemHash)
			}

			jRoot := htmlwriter.CreateHtmlNode("div").Class("form-grp")
			jRoot.Attr("formid", form.Id)
			jRoot.Attr("position", position)
			jRoot.Attr("rowindex", convert.String(rowindex))
			err = form.GenerateHTML(position, jRoot)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			var buffer bytes.Buffer
			jRoot.WriteWith(&buffer, "\t")

			html := buffer.String()
			if true {
				doc, err := goquery.NewDocumentFromReader(&buffer)
				if err != nil {
					return http.StatusInternalServerError, err
				}

				//apply data to html
				jFormGrp := doc.Find(".form-grp")
				for _, data := range dataList {
					if len(data.Value) > 0 || len(data.CodeId) > 0 {
						item, has := itemHash[data.ItemId]
						if !has {
							return http.StatusInternalServerError, fmt.Errorf("no item : %v", data.ItemId)
						}

						formKey := data.ItemId
						jTarget := jFormGrp.Find(fmt.Sprintf("[name='%s']", formKey))
						if jTarget.Length() == 0 {
							return c.JSON(http.StatusInternalServerError, &Result{Error: fmt.Errorf("target length is zero")})
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
							jTarget.Filter(fmt.Sprintf("[value='%s']", Value)).SetAttr("checked", "true")
						case "textarea":
							jTarget.AppendHtml(Value)
						default:
							jTarget.SetAttr("value", Value)
						}
					}
				}

				firstDate := convert.Time(subject.FirstDate)
				doc.Find("[name='i-1']").SetAttr("value", convert.String(firstDate.Add(time.Hour * time.Duration(24*(convert.Int(position)-1))))[:10])

				ret, err := doc.Html()
				if err != nil {
					return http.StatusInternalServerError, err
				}

				html = ret
			}

			args["FormHtml"] = template.HTML(html)
			return http.StatusOK, subject
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
				ArmId     string `json:"arm_id"`
				FirstDate string `json:"first_date"`
			}
			err := util.BodyToStruct(c.Request().Body, &item)
			if err != nil {
				return http.StatusBadRequest, err
			}
			//////////////////////////////////////////////////

			_, err = gAPI.SubjectTable.Insert(gConfig.StudyId, item.Name, item.ScrNo, item.Sex, item.BirthDate, item.ArmId, item.FirstDate, time.Now(), Uid)
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
}

type DAO_Search_Subject struct {
	Id       string `json:"id,omitempty"`
	ScrNo    string `json:"scrno"`
	Age      string `json:"age"`
	Sex      string `json:"sex"`
	Progress string `json:"progress"`
	TCreate  string `json:"t_create"`
}

func Search_Subjects(name string, scrno string, arm_id string) ([]*DAO_Search_Subject, error) {
	list, err := gAPI.SubjectTable.List(gConfig.StudyId)
	if err != nil {
		return nil, err
	}

	daos := make([]*DAO_Search_Subject, 0)
	for _, v := range list {
		if len(name) > 0 {
			if !strings.Contains(v.Name, name) {
				continue
			}
		}
		if len(scrno) > 0 {
			if !strings.Contains(v.ScrNo, scrno) {
				continue
			}
		}
		if len(arm_id) > 0 {
			if !strings.Contains(v.ArmId, arm_id) {
				continue
			}
		}

		dao := &DAO_Search_Subject{
			Id:       v.Id,
			ScrNo:    v.ScrNo,
			Age:      getAgeFromDate(v.BirthDate),
			Sex:      v.Sex,
			Progress: "X일차YY3%", //TEMP
			TCreate:  convert.String(v.TCreate)[:10],
		}
		daos = append(daos, dao)
	}
	return daos, nil
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

func Select_Subjects(subjectid string) (*DAO_Select_Subject, error) {
	subject, err := gAPI.SubjectTable.Subject(subjectid)
	if err != nil {
		return nil, err
	}

	if subject.StudyId != gConfig.StudyId {
		return nil, ErrNotMatchedStudyId
	}

	maxPosition := int64(0)

	stack, err := gAPI.StackTable.Stack(subjectid, "f-1") //TEMP
	if err != nil {
		if err != ErrNotExist {
			return nil, err
		}
	} else {
		visits, err := gAPI.VisitTable.List(stack.Id)
		if err != nil {
			return nil, err
		}

		for _, v := range visits {
			p := convert.Int(v.Position)
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
		Compliance:  make(map[string]int),
		BirthDate:   subject.BirthDate,
		FirstDate:   subject.FirstDate,
		MaxPosition: maxPosition,
		IsAuth:      false, //TEMP
		TCreate:     convert.String(subject.TCreate)[:10],
	}
	if dao.IsAuth {
		dao.Name = subject.Name
	}
	//TODO dao.Compliance
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
