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
			dao, err := Select_Subject(sbj)
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
	Id      string        `json:"id,omitempty"`
	Data    *subject.Data `json:"data"`
	TCreate string        `json:"t_create"`
}

func Select_Subject(subject *subject.Subject) (*DAO_Select_Subject, error) {
	dao := &DAO_Select_Subject{
		Id:      subject.Id,
		TCreate: convert.String(subject.TCreate)[:10],
	}
	if len(subject.Datas) > 0 {
		dao.Data = subject.Datas[len(subject.Datas)-1]
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
		Estimation.DangerousGroup = "high"
		Estimation.TargetLDL = 70
	} else if data.MedicalHistory.CoronaryArtery || data.MedicalHistory.AbdominalAneurysm || data.MedicalHistory.Diabetes {
		//고위험군 (LDL-C target <100mg/dl) : 병력 中 [경동맥질환, 복부동맥류, 당뇨병]
		Estimation.DangerousGroup = "middle-high"
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
			Estimation.DangerousGroup = "middle-low"
			Estimation.TargetLDL = 130
		} else {
			//위험인자<=1 (LDL-C target <160mg/dl)
			Estimation.DangerousGroup = "low"
			Estimation.TargetLDL = 160
		}
	}

	return Estimation, nil
}
