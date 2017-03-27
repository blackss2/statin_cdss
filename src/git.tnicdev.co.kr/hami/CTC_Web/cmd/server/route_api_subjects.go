package main

import (
	"net/http"
	"time"

	"git.tnicdev.co.kr/hami/CTC_Web/pkg/util"

	"github.com/labstack/echo"
)

func route_api_subjects(g *echo.Group) {
	g.POST("/subjects", func(c echo.Context) error {
		Uid := c.Get(UID_KEY).(string)

		retCode, retValue := (func() (int, interface{}) {
			var item struct {
				Name            string `json:"name"`
				ScrNo           string `json:"scrno"`
				Sex             string `json:"sex"`
				BirthDate       string `json:"birth_date"`
				ArmId           string `json:"arm_id"`
				InoculationDate string `json:"inoculation_date"`
			}
			err := util.BodyToStruct(c.Request().Body, &item)
			if err != nil {
				return http.StatusBadRequest, err
			}
			//////////////////////////////////////////////////

			_, err = gAPI.SubjectTable.Insert(gConfig.StudyId, item.Name, item.ScrNo, item.Sex, item.BirthDate, item.ArmId, item.InoculationDate, time.Now(), Uid)
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
