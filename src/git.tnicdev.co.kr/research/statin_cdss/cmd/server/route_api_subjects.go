package main

import (
	"net/http"
	"sort"
	"strings"
	"time"

	"git.tnicdev.co.kr/research/statin_cdss/pkg/subject"
	"git.tnicdev.co.kr/research/statin_cdss/pkg/util"

	"github.com/blackss2/utility/convert"
	"github.com/labstack/echo"
)

func route_api_subjects(g *echo.Group) {
	g.GET("/subjects", func(c echo.Context) error {
		Uid := c.Get(UID_KEY).(string)

		retCode, retValue := (func() (int, interface{}) {
			subjectid, err := util.ParamToString(c, "subjectid", false)
			if err != nil {
				return http.StatusBadRequest, err
			}
			sex, err := util.ParamToString(c, "sex", false)
			if err != nil {
				return http.StatusBadRequest, err
			}
			target_ldl, err := util.ParamToString(c, "target_ldl", false)
			if err != nil {
				return http.StatusBadRequest, err
			}
			prescription, err := util.ParamToString(c, "prescription", false)
			if err != nil {
				return http.StatusBadRequest, err
			}
			//////////////////////////////////////////////////

			subjects, err := Search_Subjects(Uid, subjectid, sex, target_ldl, prescription)
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
		Uid := c.Get(UID_KEY).(string)

		retCode, retValue := (func() (int, interface{}) {
			subjectid, err := util.PathToString(c, "subjectid")
			if err != nil {
				return http.StatusBadRequest, err
			}
			//////////////////////////////////////////////////

			subject, err := gAPI.SubjectStore.GetBySubjectId(subjectid, Uid)
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
	g.POST("/subjects", func(c echo.Context) error {
		Uid := c.Get(UID_KEY).(string)

		retCode, retValue := (func() (int, interface{}) {
			var item struct {
				SubjectId string `json:"subject_id"`
			}
			err := util.BodyToStruct(c.Request().Body, &item)
			if err != nil {
				return http.StatusBadRequest, err
			}
			//////////////////////////////////////////////////

			_, err = gAPI.SubjectStore.Insert(item.SubjectId, Uid, time.Now())
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
		Uid := c.Get(UID_KEY).(string)

		retCode, retValue := (func() (int, interface{}) {
			subjectid, err := util.PathToString(c, "subjectid")
			if err != nil {
				return http.StatusBadRequest, err
			}

			var item struct {
				//TODO
			}
			err = util.BodyToStruct(c.Request().Body, &item)
			if err != nil {
				return http.StatusBadRequest, err
			}
			//////////////////////////////////////////////////

			subject, err := gAPI.SubjectStore.GetBySubjectId(subjectid, Uid)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			//TODO

			err = gAPI.SubjectStore.Update(subject.Id, subject.SubjectId) //TEMP
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
	Id           string `json:"id,omitempty"`
	SubjectId    string `json:"subject_id"`
	Sex          string `json:"sex"`
	TargetLDL    string `json:"target_ldl"`
	Prescription string `json:"prescription"`
	TCreate      string `json:"t_create"`
}

func Search_Subjects(Uid string, SubjectId string, Sex string, TargetLDL string, Prescription string) ([]*DAO_Search_Subject, error) {
	list, err := gAPI.SubjectStore.ListByOwnerId(Uid)
	if err != nil {
		return nil, err
	}

	daos := make([]*DAO_Search_Subject, 0)
	for _, v := range list {
		if len(SubjectId) > 0 {
			if !strings.Contains(v.SubjectId, SubjectId) {
				continue
			}
		}

		//TODO : Sex, TargetLDL, Prescription

		dao := &DAO_Search_Subject{
			Id:        v.Id,
			SubjectId: v.SubjectId,
			//Sex:        v.Sex,
			//TargetLDL:        v.TargetLDL,
			//Prescription:        v.Prescription,
			TCreate: convert.String(v.TCreate)[:10],
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
	return s[i].SubjectId < s[j].SubjectId
}

type DAO_Select_Subject struct {
	Id string `json:"id,omitempty"`
	//TODO
	TCreate string `json:"t_create"`
}

func Select_Subject(subject *subject.Subject) (*DAO_Select_Subject, error) {
	dao := &DAO_Select_Subject{
		Id: subject.Id,
		//TODO
		TCreate: convert.String(subject.TCreate)[:10],
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
