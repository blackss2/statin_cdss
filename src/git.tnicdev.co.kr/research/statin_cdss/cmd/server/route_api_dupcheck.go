package main

import (
	"net/http"

	"git.tnicdev.co.kr/research/statin_cdss/pkg/subject"
	"git.tnicdev.co.kr/research/statin_cdss/pkg/user"
	"git.tnicdev.co.kr/research/statin_cdss/pkg/util"

	"github.com/labstack/echo"
)

func route_api_dupcheck(g *echo.Group) {
	g.GET("/dupcheck/users/:userid", func(c echo.Context) error {
		retCode, retValue := (func() (int, interface{}) {
			userid, err := util.PathToString(c, "userid")
			if err != nil {
				return http.StatusBadRequest, err
			}
			//////////////////////////////////////////////////

			_, err = gAPI.UserStore.GetByUserId(userid)
			if err != nil {
				if err != user.ErrExistUserId {
					return http.StatusInternalServerError, err
				}
				return http.StatusOK, true
			} else {
				return http.StatusOK, false
			}
		})()

		if err, is := retValue.(error); is {
			return c.JSON(retCode, &Result{Error: err})
		} else {
			return c.JSON(retCode, &Result{Result: retValue})
		}
	})
	g.GET("/dupcheck/subjects/:subjectid", func(c echo.Context) error {
		Uid := c.Get(UID_KEY).(string)

		retCode, retValue := (func() (int, interface{}) {
			subjectid, err := util.PathToString(c, "subjectid")
			if err != nil {
				return http.StatusBadRequest, err
			}
			//////////////////////////////////////////////////

			_, err = gAPI.SubjectStore.GetBySubjectId(subjectid, Uid)
			if err != nil {
				if err != subject.ErrExistSubjectId {
					return http.StatusInternalServerError, err
				}
				return http.StatusOK, true
			} else {
				return http.StatusOK, false
			}
		})()

		if err, is := retValue.(error); is {
			return c.JSON(retCode, &Result{Error: err})
		} else {
			return c.JSON(retCode, &Result{Result: retValue})
		}
	})
}
