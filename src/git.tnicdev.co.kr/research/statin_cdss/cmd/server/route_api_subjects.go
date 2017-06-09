package main

import (
	"errors"
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

			sbj, err := gAPI.SubjectStore.GetBySubjectId(subjectid, Uid)
			if err != nil {
				return http.StatusInternalServerError, err
			}
			dao, err := Select_Subject(Uid, sbj)
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
				SubjectId string `json:"subjectid"`
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
	g.POST("/subjects/:subjectid/initial", func(c echo.Context) error {
		Uid := c.Get(UID_KEY).(string)

		retCode, retValue := (func() (int, interface{}) {
			subjectid, err := util.PathToString(c, "subjectid")
			if err != nil {
				return http.StatusBadRequest, err
			}

			var item struct {
				Demography     subject.Demography     `json:"demography"`
				BloodPressure  subject.BloodPressure  `json:"blood_pressure"`
				StatinFirst    subject.StatinFirst    `json:"statin_first"`
				StatinsLast    subject.StatinsLast    `json:"statin_last"`
				BloodTest      subject.BloodTest      `json:"blood_test"`
				MedicalHistory subject.MedicalHistory `json:"medical_history"`
				FamilyHistory  subject.FamilyHistory  `json:"family_history"`
			}
			err = util.BodyToStruct(c.Request().Body, &item)
			if err != nil {
				return http.StatusBadRequest, err
			}
			//////////////////////////////////////////////////

			sbj, err := gAPI.SubjectStore.GetBySubjectId(subjectid, Uid)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			if len(sbj.Datas) > 0 {
				return http.StatusBadRequest, errors.New("exist initial data")
			}

			data := &subject.Data{
				Demography:     item.Demography,
				BloodPressure:  item.BloodPressure,
				StatinFirst:    item.StatinFirst,
				StatinsLast:    item.StatinsLast,
				BloodTest:      item.BloodTest,
				MedicalHistory: item.MedicalHistory,
				FamilyHistory:  item.FamilyHistory,
				TCreate:        time.Now(),
			}
			Estimation, err := CalculateEstimation(data)
			data.Estimation = Estimation

			err = gAPI.SubjectStore.AppendData(sbj.Id, data)
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
	g.POST("/subjects/:subjectid/followup", func(c echo.Context) error {
		Uid := c.Get(UID_KEY).(string)

		retCode, retValue := (func() (int, interface{}) {
			subjectid, err := util.PathToString(c, "subjectid")
			if err != nil {
				return http.StatusBadRequest, err
			}

			var item struct {
				Demography     subject.Demography     `json:"demography"`
				BloodPressure  subject.BloodPressure  `json:"blood_pressure"`
				StatinFirst    subject.StatinFirst    `json:"statin_first"`
				StatinsLast    subject.StatinsLast    `json:"statin_last"`
				BloodTest      subject.BloodTest      `json:"blood_test"`
				MedicalHistory subject.MedicalHistory `json:"medical_history"`
				FamilyHistory  subject.FamilyHistory  `json:"family_history"`
			}
			err = util.BodyToStruct(c.Request().Body, &item)
			if err != nil {
				return http.StatusBadRequest, err
			}
			//////////////////////////////////////////////////

			sbj, err := gAPI.SubjectStore.GetBySubjectId(subjectid, Uid)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			if len(sbj.Datas) == 0 {
				return http.StatusBadRequest, errors.New("not exist initial data")
			}

			last := sbj.Datas[len(sbj.Datas)-1]
			item.Demography.Age = last.Demography.Age
			item.Demography.BirthDate = last.Demography.BirthDate
			item.Demography.Height = last.Demography.Height
			item.Demography.Sex = last.Demography.Sex
			if item.Demography.Weight == 0 {
				item.Demography.Weight = last.Demography.Weight
			}
			if len(item.BloodPressure.Date) == 0 || item.BloodPressure.Systolic == 0 || item.BloodPressure.Diastolic == 0 {
				item.BloodPressure = last.BloodPressure
			}
			item.StatinFirst = last.StatinFirst
			if len(item.StatinsLast.Dept) == 0 || len(item.StatinsLast.Code) == 0 || len(item.StatinsLast.Date) == 0 || item.StatinsLast.Period == 0 {
				item.StatinsLast = last.StatinsLast
			}
			if len(item.BloodTest.Date) == 0 || item.BloodTest.HDL == 0 || item.BloodTest.TotalCholesterol == 0 || item.BloodTest.Glucose == 0 {
				item.BloodTest = last.BloodTest
			}
			if last.MedicalHistory.TransientStroke {
				item.MedicalHistory.TransientStroke = last.MedicalHistory.TransientStroke
			}
			if last.MedicalHistory.PeripheralVascular {
				item.MedicalHistory.PeripheralVascular = last.MedicalHistory.PeripheralVascular
			}
			if last.MedicalHistory.Carotid {
				item.MedicalHistory.Carotid = last.MedicalHistory.Carotid
			}
			if last.MedicalHistory.AbdominalAneurysm {
				item.MedicalHistory.AbdominalAneurysm = last.MedicalHistory.AbdominalAneurysm
			}
			if last.MedicalHistory.Diabetes {
				item.MedicalHistory.Diabetes = last.MedicalHistory.Diabetes
			}
			if last.MedicalHistory.CoronaryArtery {
				item.MedicalHistory.CoronaryArtery = last.MedicalHistory.CoronaryArtery
			}
			if last.MedicalHistory.IschemicStroke {
				item.MedicalHistory.IschemicStroke = last.MedicalHistory.IschemicStroke
			}
			if last.MedicalHistory.HighBloodPressure {
				item.MedicalHistory.HighBloodPressure = last.MedicalHistory.HighBloodPressure
			}
			if last.MedicalHistory.Smoking {
				item.MedicalHistory.Smoking = last.MedicalHistory.Smoking
			}
			if last.FamilyHistory.CoronaryArtery {
				item.FamilyHistory.CoronaryArtery = last.FamilyHistory.CoronaryArtery
			}

			data := &subject.Data{
				Demography:     item.Demography,
				BloodPressure:  item.BloodPressure,
				StatinFirst:    item.StatinFirst,
				StatinsLast:    item.StatinsLast,
				BloodTest:      item.BloodTest,
				MedicalHistory: item.MedicalHistory,
				FamilyHistory:  item.FamilyHistory,
				TCreate:        time.Now(),
			}
			Estimation, err := CalculateEstimation(data)
			data.Estimation = Estimation

			err = gAPI.SubjectStore.AppendData(sbj.Id, data)
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
	g.PUT("/subjects/:subjectid/share", func(c echo.Context) error {
		Uid := c.Get(UID_KEY).(string)

		retCode, retValue := (func() (int, interface{}) {
			subjectid, err := util.PathToString(c, "subjectid")
			if err != nil {
				return http.StatusBadRequest, err
			}

			var item struct {
				Share bool `json:"share"`
			}
			err = util.BodyToStruct(c.Request().Body, &item)
			if err != nil {
				return http.StatusBadRequest, err
			}
			//////////////////////////////////////////////////

			sbj, err := gAPI.SubjectStore.GetBySubjectId(subjectid, Uid)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			err = gAPI.SubjectStore.SetShare(sbj.Id, item.Share)
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

	g.POST("/subjects/:subjectid/prescription", func(c echo.Context) error {
		Uid := c.Get(UID_KEY).(string)

		retCode, retValue := (func() (int, interface{}) {
			subjectid, err := util.PathToString(c, "subjectid")
			if err != nil {
				return http.StatusBadRequest, err
			}

			var item struct {
				Prescription subject.Prescription `json:"prescription"`
			}
			err = util.BodyToStruct(c.Request().Body, &item)
			if err != nil {
				return http.StatusBadRequest, err
			}
			//////////////////////////////////////////////////

			sbj, err := gAPI.SubjectStore.GetBySubjectId(subjectid, Uid)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			if len(sbj.Datas) == 0 {
				return http.StatusBadRequest, errors.New("not exist initial data")
			}

			last := sbj.Datas[len(sbj.Datas)-1]

			if len(last.Prescription.Statins) > 0 || len(last.Prescription.Levels) > 0 {
				data := &subject.Data{
					Demography:     last.Demography,
					BloodPressure:  last.BloodPressure,
					StatinFirst:    last.StatinFirst,
					StatinsLast:    last.StatinsLast,
					BloodTest:      last.BloodTest,
					MedicalHistory: last.MedicalHistory,
					FamilyHistory:  last.FamilyHistory,
					Estimation:     last.Estimation,
					Prescription:   item.Prescription,
					TCreate:        time.Now(),
				}

				err = gAPI.SubjectStore.AppendData(sbj.Id, data)
				if err != nil {
					return http.StatusInternalServerError, err
				}
			} else {
				last.Prescription = item.Prescription

				err = gAPI.SubjectStore.UpdateLastData(sbj.Id, last)
				if err != nil {
					return http.StatusInternalServerError, err
				}
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
	Id             string `json:"id,omitempty"`
	SubjectId      string `json:"subject_id"`
	Share          bool   `json:"share"`
	Own            bool   `json:"own"`
	Sex            string `json:"sex"`
	TargetLDL      string `json:"target_ldl"`
	DangerousGroup string `json:"dangerous_group"`
	TCreate        string `json:"t_create"`
}

func Search_Subjects(Uid string, SubjectId string, Sex string, TargetLDL string, Prescription string) ([]*DAO_Search_Subject, error) {
	list, err := gAPI.SubjectStore.ListByOwnerId(Uid)
	if err != nil {
		return nil, err
	}
	list_share, err := gAPI.SubjectStore.ListShare()
	if err != nil {
		return nil, err
	}
	for _, v := range list_share {
		if v.OwnerId != Uid {
			list = append(list, v)
		}
	}

	daos := make([]*DAO_Search_Subject, 0)
	for _, v := range list {
		if len(SubjectId) > 0 {
			if !strings.Contains(v.SubjectId, SubjectId) {
				continue
			}
		}

		dao := &DAO_Search_Subject{
			Id:        v.Id,
			SubjectId: v.SubjectId,
			Share:     v.Share,
			Own:       v.OwnerId == Uid,
			TCreate:   convert.String(v.TCreate)[:10],
		}

		if len(v.Datas) > 0 {
			data := v.Datas[len(v.Datas)-1]
			dao.Sex = data.Demography.Sex
			dao.DangerousGroup = data.Estimation.DangerousGroup
			dao.TargetLDL = convert.String(data.Estimation.TargetLDL)
		}

		if len(Sex) > 0 {
			if Sex != dao.Sex {
				continue
			}
		}
		if len(TargetLDL) > 0 {
			if TargetLDL != dao.TargetLDL {
				continue
			}
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
	Id        string         `json:"id,omitempty"`
	SubjectId string         `json:"subject_id"`
	Share     bool           `json:"share"`
	Own       bool           `json:"own"`
	Data      *subject.Data  `json:"data"`
	History   []*DAO_History `json:"history"`
	TCreate   string         `json:"t_create"`
}

type DAO_History struct {
	Height            float64  `json:"height"`
	Weight            float64  `json:"weight"`
	Systolic          int64    `json:"systolic"`
	Diastolic         int64    `json:"diastolic"`
	HDL               float64  `json:"hdl"`
	TotalCholesterol  float64  `json:"total_cholesterol"`
	Glucose           float64  `json:"glucose"`
	Diabetes          bool     `json:"diabetes"`
	HighBloodPressure bool     `json:"high_blood_pressure"`
	DangerousGroup    string   `json:"dangerous_group"`
	TargetLDL         float64  `json:"target_ldl"`
	Statins           []string `json:"statins"`
	Levels            []string `json:"levels"`
	TCreate           string   `json:"t_create"`
}

func Select_Subject(Uid string, subject *subject.Subject) (*DAO_Select_Subject, error) {
	dao := &DAO_Select_Subject{
		Id:        subject.Id,
		SubjectId: subject.SubjectId,
		Share:     subject.Share,
		Own:       subject.OwnerId == Uid,
		History:   make([]*DAO_History, 0),
		TCreate:   convert.String(subject.TCreate)[:10],
	}
	if len(subject.Datas) > 0 {
		dao.Data = subject.Datas[len(subject.Datas)-1]
	}
	for i, _ := range subject.Datas {
		v := subject.Datas[len(subject.Datas)-i-1]
		h := &DAO_History{
			Height:            v.Demography.Height,
			Weight:            v.Demography.Weight,
			Systolic:          v.BloodPressure.Systolic,
			Diastolic:         v.BloodPressure.Diastolic,
			HDL:               v.BloodTest.HDL,
			TotalCholesterol:  v.BloodTest.TotalCholesterol,
			Glucose:           v.BloodTest.Glucose,
			Diabetes:          v.MedicalHistory.Diabetes,
			HighBloodPressure: v.MedicalHistory.HighBloodPressure,
			DangerousGroup:    v.Estimation.DangerousGroup,
			TargetLDL:         v.Estimation.TargetLDL,
			Statins:           v.Prescription.Statins,
			Levels:            v.Prescription.Levels,
			TCreate:           convert.String(v.TCreate)[:10],
		}
		dao.History = append(dao.History, h)
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

func CalculateEstimation(data *subject.Data) (subject.Estimation, error) {
	var Estimation subject.Estimation
	if data.MedicalHistory.CoronaryArtery || data.MedicalHistory.IschemicStroke || data.MedicalHistory.TransientStroke || data.MedicalHistory.PeripheralVascular {
		//초고위험군 (LDL-C target <70mg/dl) : 병력 中 [관상동맥질환, 허혈성 뇌졸증, 일과성 뇌허혈발작, 말초혈관질환]
		Estimation.DangerousGroup = "extream"
		Estimation.TargetLDL = 70
	} else if data.MedicalHistory.CoronaryArtery || data.MedicalHistory.AbdominalAneurysm || data.MedicalHistory.Diabetes {
		//고위험군 (LDL-C target <100mg/dl) : 병력 中 [경동맥질환, 복부동맥류, 당뇨병]
		Estimation.DangerousGroup = "high"
		Estimation.TargetLDL = 100
	} else {
		//위험인자 : 흡연, 고혈압(수축기>=140 OR 이완기>=90), HDL-C(<40mg/dL), HDL-C(>=60mg/dL), 연령(남>=45, 여>=55), 관상동맥질환 조기발병 가족력
		count := 0
		if data.MedicalHistory.Smoking {
			count++
		}
		if data.BloodPressure.Systolic >= 140 || data.BloodPressure.Diastolic >= 90 {
			count++
		}
		if data.BloodTest.HDL < 40 {
			count++
		}
		if data.BloodTest.HDL >= 60 {
			count++
		}
		if data.Demography.Sex == "M" && data.Demography.Age >= 45 {
			count++
		}
		if data.FamilyHistory.CoronaryArtery {
			count++
		}
		if count >= 2 {
			//위험인자>=2 (LDL-C target <130mg/dl)
			Estimation.DangerousGroup = "danger2"
			Estimation.TargetLDL = 130
		} else {
			//위험인자<=1 (LDL-C target <160mg/dl)
			Estimation.DangerousGroup = "danger1"
			Estimation.TargetLDL = 160
		}
	}

	return Estimation, nil
}
