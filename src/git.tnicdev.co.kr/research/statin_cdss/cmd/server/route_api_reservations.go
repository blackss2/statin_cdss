package main

import (
	"net/http"
	"time"

	"git.tnicdev.co.kr/research/statin_cdss/pkg/reservation"
	"git.tnicdev.co.kr/research/statin_cdss/pkg/store"
	"git.tnicdev.co.kr/research/statin_cdss/pkg/util"

	"github.com/blackss2/utility/convert"
	"github.com/labstack/echo"
)

func route_api_reservations(g *echo.Group) {
	g.GET("/reservations", func(c echo.Context) error {
		retCode, retValue := (func() (int, interface{}) {
			t_start, err := util.ParamToTime(c, "t_start", true)
			if err != nil {
				return http.StatusBadRequest, err
			}
			t_end, err := util.ParamToTime(c, "t_end", true)
			if err != nil {
				return http.StatusBadRequest, err
			}

			if !t_start.Valid {
				return http.StatusBadRequest, ErrInvalidDate
			}
			if !t_end.Valid {
				return http.StatusBadRequest, ErrInvalidDate
			}
			//////////////////////////////////////////////////

			reservations, err := Calendar_Reservations(t_start.Time, t_end.Time)
			if err != nil {
				return http.StatusInternalServerError, err
			}
			return http.StatusOK, reservations
		})()
		if err, is := retValue.(error); is {
			return c.JSON(retCode, &Result{Error: err})
		} else {
			return c.JSON(retCode, &Result{Result: retValue})
		}
	})
	g.POST("/reservations", func(c echo.Context) error {
		Uid := c.Get(UID_KEY).(string)

		retCode, retValue := (func() (int, interface{}) {
			var item struct {
				Name  string        `json:"name"`
				TDate util.NullTime `json:"t_date"`
			}
			err := util.BodyToStruct(c.Request().Body, &item)
			if err != nil {
				return http.StatusBadRequest, err
			}

			if !item.TDate.Valid {
				return http.StatusBadRequest, ErrInvalidDate
			}
			//////////////////////////////////////////////////

			_, err = gAPI.ReservationStore.Insert(gConfig.StudyId, item.Name, item.TDate.Time, time.Now(), Uid)
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
	g.PUT("/reservations/:reservationid", func(c echo.Context) error {
		retCode, retValue := (func() (int, interface{}) {
			reservationid, err := util.PathToString(c, "reservationid")
			if err != nil {
				return http.StatusBadRequest, err
			}
			//////////////////////////////////////////////////

			var item struct {
				Name string `json:"name"`
			}
			err = util.BodyToStruct(c.Request().Body, &item)
			if err != nil {
				return http.StatusBadRequest, err
			}
			//////////////////////////////////////////////////

			err = gAPI.ReservationStore.Update(reservationid, item.Name)
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
	g.DELETE("/reservations/:reservationid", func(c echo.Context) error {
		retCode, retValue := (func() (int, interface{}) {
			reservationid, err := util.PathToString(c, "reservationid")
			if err != nil {
				return http.StatusBadRequest, err
			}
			//////////////////////////////////////////////////

			err = gAPI.ReservationStore.Delete(reservationid)
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
	g.POST("/reservations/:reservationid", func(c echo.Context) error {
		retCode, retValue := (func() (int, interface{}) {
			reservationid, err := util.PathToString(c, "reservationid")
			if err != nil {
				return http.StatusBadRequest, err
			}

			var item struct {
				ScrNo   string `json:"scrno"`
				Minutes int64  `json:"minutes"`
				Status  int64  `json:"status"`
			}
			err = util.BodyToStruct(c.Request().Body, &item)
			if err != nil {
				return http.StatusBadRequest, err
			}
			//////////////////////////////////////////////////

			subject, err := gAPI.SubjectTable.SubjectByScrNo(item.ScrNo)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			err = gAPI.ReservationStore.AddSubject(reservationid, subject.Id, item.Minutes, item.Status)
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
	g.PUT("/reservations/:reservationid/subjects/:subjectid", func(c echo.Context) error {
		retCode, retValue := (func() (int, interface{}) {
			reservationid, err := util.PathToString(c, "reservationid")
			if err != nil {
				return http.StatusBadRequest, err
			}
			subjectid, err := util.PathToString(c, "subjectid")
			if err != nil {
				return http.StatusBadRequest, err
			}

			var item struct {
				Minutes int64 `json:"minutes"`
				Status  int64 `json:"status"`
			}
			err = util.BodyToStruct(c.Request().Body, &item)
			if err != nil {
				return http.StatusBadRequest, err
			}
			//////////////////////////////////////////////////

			err = gAPI.ReservationStore.UpdateSubject(reservationid, subjectid, item.Minutes, item.Status)
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
	g.DELETE("/reservations/:reservationid/subjects/:subjectid", func(c echo.Context) error {
		retCode, retValue := (func() (int, interface{}) {
			reservationid, err := util.PathToString(c, "reservationid")
			if err != nil {
				return http.StatusBadRequest, err
			}
			subjectid, err := util.PathToString(c, "subjectid")
			if err != nil {
				return http.StatusBadRequest, err
			}
			//////////////////////////////////////////////////

			err = gAPI.ReservationStore.DeleteSubject(reservationid, subjectid)
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

type DAO_Calendar_Reservation struct {
	Id       string                             `json:"id,omitempty"`
	Name     string                             `json:"name"`
	Date     string                             `json:"date"`
	Subjects []*DAO_Calendar_ReservationSubject `json:"subjects"`
}

type DAO_Calendar_ReservationSubject struct {
	SubjectId string `json:"subject_id"`
	Minutes   int64  `json:"minutes"`
	ScrNo     string `json:"scrno"`
	Status    int64  `json:"status"`
}

func Calendar_Reservations(TStart time.Time, TEnd time.Time) ([]*DAO_Calendar_Reservation, error) {
	var list []*reservation.Reservation
	err := gAPI.ReservationStore.List(&list, store.ListOption{
		WhereOption: store.WhereOption{
			IndexBy:       "study_id",
			IndexByValues: []interface{}{gConfig.StudyId},
		},
	})
	if err != nil {
		return nil, err
	}

	subjects, err := gAPI.SubjectTable.List(gConfig.StudyId)
	if err != nil {
		return nil, err
	}

	subjectHash := make(map[string]*Subject)
	for _, v := range subjects {
		subjectHash[v.Id] = v
	}

	daos := make([]*DAO_Calendar_Reservation, 0)
	for _, v := range list {
		if v.TDate.Before(TStart) || v.TDate.After(TEnd) {
			continue
		}
		dao := &DAO_Calendar_Reservation{
			Id:       v.Id,
			Name:     v.Name,
			Date:     convert.String(v.TDate)[:10],
			Subjects: make([]*DAO_Calendar_ReservationSubject, 0),
		}
		for _, s := range v.Subjects {
			subject, has := subjectHash[s.SubjectId]
			if !has {
				return nil, ErrNotExistSubject
			}
			dao.Subjects = append(dao.Subjects, &DAO_Calendar_ReservationSubject{
				SubjectId: s.SubjectId,
				ScrNo:     subject.ScrNo,
				Minutes:   s.Minutes,
				Status:    s.Status,
			})
		}
		daos = append(daos, dao)
	}
	return daos, nil
}
