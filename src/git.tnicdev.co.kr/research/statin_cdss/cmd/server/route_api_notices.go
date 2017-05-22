package main

import (
	"net/http"
	"time"

	"git.tnicdev.co.kr/research/statin_cdss/pkg/notice"
	"git.tnicdev.co.kr/research/statin_cdss/pkg/store"
	"git.tnicdev.co.kr/research/statin_cdss/pkg/util"

	"github.com/labstack/echo"
)

func route_api_notices(g *echo.Group) {
	g.GET("/notices", func(c echo.Context) error {
		retCode, retValue := (func() (int, interface{}) {
			offset, err := util.ParamToInt(c, "offset", false)
			if err != nil {
				return http.StatusBadRequest, err
			}
			limit, err := util.ParamToInt(c, "limit", false)
			if err != nil {
				return http.StatusBadRequest, err
			}
			//////////////////////////////////////////////////

			notices, notice_count, err := Search_Notices(offset, limit)
			if err != nil {
				return http.StatusInternalServerError, err
			}
			return http.StatusOK, map[string]interface{}{
				"notices":      notices,
				"notice_count": notice_count,
			}
		})()
		if err, is := retValue.(error); is {
			return c.JSON(retCode, &Result{Error: err})
		} else {
			return c.JSON(retCode, &Result{Result: retValue})
		}
	})
	g.POST("/notices", func(c echo.Context) error {
		Uid := c.Get(UID_KEY).(string)

		retCode, retValue := (func() (int, interface{}) {
			var item struct {
				Content string `json:"content"`
			}
			err := util.BodyToStruct(c.Request().Body, &item)
			if err != nil {
				return http.StatusBadRequest, err
			}
			//////////////////////////////////////////////////

			_, err = gAPI.NoticeStore.Insert(gConfig.StudyId, item.Content, time.Now(), Uid)
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
	g.DELETE("/notices/:noticeid", func(c echo.Context) error {
		retCode, retValue := (func() (int, interface{}) {
			noticeid, err := util.PathToString(c, "noticeid")
			if err != nil {
				return http.StatusBadRequest, err
			}
			//////////////////////////////////////////////////

			err = gAPI.NoticeStore.Delete(noticeid)
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
	g.POST("/notice_push", func(c echo.Context) error {
		retCode, retValue := (func() (int, interface{}) {
			//TOOD
			return http.StatusOK, true
		})()
		if err, is := retValue.(error); is {
			return c.JSON(retCode, &Result{Error: err})
		} else {
			return c.JSON(retCode, &Result{Result: retValue})
		}
	})
}

type DAO_Search_Notice struct {
	Id      string `json:"id"`
	Content string `json:"content"`
}

func Search_Notices(offset int64, limit int64) ([]*DAO_Search_Notice, int64, error) {
	whereOpt := store.WhereOption{
		IndexBy:       "study_id",
		IndexByValues: []interface{}{gConfig.StudyId},
	}

	var list []*notice.Notice
	err := gAPI.NoticeStore.List(&list, store.ListOption{
		WhereOption: whereOpt,
		Offset:      int(offset),
		Limit:       int(limit),
	})
	if err != nil {
		return nil, 0, err
	}

	count, err := gAPI.NoticeStore.Count(store.ListOption{
		WhereOption: whereOpt,
	})
	if err != nil {
		return nil, 0, err
	}

	daos := make([]*DAO_Search_Notice, 0)
	for _, v := range list {
		dao := &DAO_Search_Notice{
			Id:      v.Id,
			Content: v.Content,
		}
		daos = append(daos, dao)
	}
	return daos, count, nil
}
