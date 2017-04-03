package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"git.tnicdev.co.kr/hami/CTC_Web/pkg/reservation"
	"git.tnicdev.co.kr/hami/CTC_Web/pkg/user"
	"git.tnicdev.co.kr/hami/CTC_Web/pkg/util"

	"github.com/labstack/echo"
)

func main_web(wg sync.WaitGroup) {
	defer wg.Done()

	PORT := gConfig.Port.Web

	e := echo.New()
	//API
	g := e.Group("/api")
	g.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("Cache-Control", "no-cache")
			return next(c)
		}
	})
	route_api_subjects(g)
	route_api_notices(g)
	route_api_reservations(g)

	//Web Pages
	web := util.NewWebServer(e, "./webfiles")
	StaticFiles := make(map[string]*util.StaticFilesFile)
	for k, v := range staticFiles {
		StaticFiles[k] = &util.StaticFilesFile{
			Data:  v.data,
			Mime:  v.mime,
			Mtime: v.mtime,
			Size:  v.size,
			Hash:  v.hash,
		}
	}
	util.SetStaticFiles(StaticFiles)

	e.Renderer = web
	web.SetupStatic(e, "/public", "./webfiles/public")

	webChecker := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			web.CheckWatch()
			return next(c)
		}
	}
	route_login(e, webChecker)

	e.GET("/", func(c echo.Context) error {
		args := make(map[string]interface{})
		_, err := InitUserArgs(c, args)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, &Result{Error: err})
		}

		if args["IsLogin"] == true {
			subjects, err := gAPI.SubjectTable.List(gConfig.StudyId)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, &Result{Error: err})
			}
			subjectHash := make(map[string]*Subject)
			for _, v := range subjects {
				subjectHash[v.Id] = v
			}

			total_ae := make(map[string]int)
			arm_ae := make(map[string]int)
			if true {
				subjectAEHash := make(map[string]bool)
				armAEHash := make(map[string]int)
				for _, v := range subjects {
					count, err := Subject_AECount(v.Id)
					if err != nil {
						return c.JSON(http.StatusInternalServerError, &Result{Error: err})
					}
					if count > 0 {
						subjectAEHash[v.Id] = true
						armAEHash[v.Arm] = count
					}
				}

				total_ae["정상"] = len(subjects) - len(subjectAEHash)
				total_ae["이상반응"] = len(subjectAEHash)

				arm_ae["실험군"] = armAEHash["실험군"]
				arm_ae["대조군"] = armAEHash["대조군"]
				arm_ae["위약군"] = armAEHash["위약군"]
			}

			compliances := make([]*DAO_Compliance, 0)

			DAY_COUNT := 7

			cpHash := make(map[string]*Compliance)
			for _, v := range subjects {
				cps, err := Subject_Compliance(v.Id)
				if err != nil {
					return c.JSON(http.StatusInternalServerError, &Result{Error: err})
				}
				for _, v := range cps {
					cp, has := cpHash[v.Day]
					if !has {
						cp = new(Compliance)
						cpHash[v.Day] = cp
					}
					cp.LocalCount += v.LocalCount
					cp.LocalItemCount += v.LocalItemCount
					cp.SysmeticCount += v.SysmeticCount
					cp.SysmeticItemCount += v.SysmeticItemCount
					cp.VitalCount += v.VitalCount
					cp.VitalItemCount += v.VitalItemCount
					cp.TotalCount += v.TotalCount
					cp.TotalItemCount += v.TotalItemCount
				}
			}

			for i := 0; i < DAY_COUNT; i++ {
				p := fmt.Sprintf("%d", i+1)
				cp := cpHash[p]
				dao := &DAO_Compliance{
					Day: p,
				}
				if cp.LocalItemCount > 0 {
					dao.Local = cp.LocalCount * 100 / cp.LocalItemCount
				}
				if cp.SysmeticItemCount > 0 {
					dao.Sysmetic = cp.SysmeticCount * 100 / cp.SysmeticItemCount
				}
				if cp.VitalItemCount > 0 {
					dao.Vital = cp.VitalCount * 100 / cp.VitalItemCount
				}
				if cp.TotalItemCount > 0 {
					dao.Total = cp.TotalCount * 100 / cp.TotalItemCount
				}

				compliances = append(compliances, dao)
			}

			//TODO
			enrollment := make(map[string]int)
			//TODO
			reservations := make(map[string]int)
			if true {
				var list []*reservation.Reservation
				err := gAPI.ReservationStore.List(&list)
				if err != nil {
					return c.JSON(http.StatusInternalServerError, &Result{Error: err})
				}

				reservations["실험군"] = 0
				reservations["대조군"] = 0
				reservations["위약군"] = 0
				for _, v := range list {
					for _, s := range v.Subjects {
						if subject, has := subjectHash[s.SubjectId]; has {
							reservations[subject.Arm]++
							reservations["total"]++
						}
					}
				}
			}

			notices, notice_count, err := Search_Notices(0, 5)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, &Result{Error: err})
			}

			args["page_initial"] = map[string]interface{}{
				"total_ae":     total_ae,
				"arm_ae":       arm_ae,
				"compliances":  compliances,
				"enrollment":   enrollment,
				"reservations": reservations,
				"notices":      notices,
				"notice_count": notice_count,
			}
			return c.Render(http.StatusOK, "main.html", args)
		} else {
			args["page_initial"] = map[string]interface{}{}
			return c.Render(http.StatusOK, "login.html", args)
		}
	}, webChecker)

	e.GET("/subject", func(c echo.Context) error {
		args := make(map[string]interface{})
		user, err := InitUserArgs(c, args)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, &Result{Error: err})
		}

		args["page_initial"] = map[string]interface{}{}

		InitSidebarArgs(c, user, args)
		args["has_sidebar"] = true
		return c.Render(http.StatusOK, "subject.html", args)
	}, webChecker)

	e.GET("/schedule", func(c echo.Context) error {
		args := make(map[string]interface{})
		_, err := InitUserArgs(c, args)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, &Result{Error: err})
		}
		args["page_initial"] = map[string]interface{}{}
		return c.Render(http.StatusOK, "schedule.html", args)
	}, webChecker)

	e.GET("/export", func(c echo.Context) error {
		args := make(map[string]interface{})
		_, err := InitUserArgs(c, args)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, &Result{Error: err})
		}

		args["page_initial"] = map[string]interface{}{}
		return c.Render(http.StatusOK, "export.html", args)
	}, webChecker)

	e.GET("/:filename", func(c echo.Context) error {
		args := make(map[string]interface{})
		InitUserArgs(c, args)
		filename := c.Param("filename")
		return c.Render(http.StatusOK, filename, args)
	}, webChecker)

	e.GET("/favicon.ico", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	}, webChecker)
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		log.Println(err, c.Request().URL.Path)
		c.String(c.Response().Status, err.Error())
	}

	if gConfig.SSL.Enable {
		if len(gConfig.SSL.RedirectPort) > 0 {
			pe := echo.New()
			pe.Any("/*", func(c echo.Context) error {
				Hostname := strings.Split(c.Request().Host, ":")[0]
				var buffer bytes.Buffer
				buffer.WriteString("https://")
				buffer.WriteString(Hostname)
				if PORT != "443" {
					buffer.WriteString(":")
					buffer.WriteString(PORT)
				}
				if len(c.Request().URL.Path) > 0 {
					buffer.WriteString(c.Request().URL.Path)
				}
				if len(c.Request().URL.RawQuery) > 0 {
					buffer.WriteString("?")
					buffer.WriteString(c.Request().URL.RawQuery)
				}
				return c.Redirect(http.StatusFound, buffer.String())
			}, webChecker)
			go pe.Start(":" + gConfig.SSL.RedirectPort)
		}
		log.Println(e.StartTLS(":"+PORT, gConfig.SSL.Cert.Public, gConfig.SSL.Cert.Private))
	} else {
		log.Println(e.Start(":" + PORT))
	}
}

type DAO_Common_UserInfo struct {
	UserId       string `json:"userid"`
	Name         string `json:"name"`
	Birth        string `json:"birth"`
	Mobile       string `json:"mobile"`
	Organization string `json:"organization"`
	Position     string `json:"position"`
	Role         string `json:"role"`
}

func InitUserArgs(c echo.Context, args map[string]interface{}) (*user.User, error) {
	Uid := c.Get(UID_KEY)
	args["IsLogin"] = (Uid != nil)
	if Uid != nil {
		var user *user.User
		err := gAPI.UserStore.Get(Uid.(string), &user)
		if err != nil {
			return nil, err
		}
		args["uid"] = user.Id
		args["role"] = user.Role

		var dao = &DAO_Common_UserInfo{
			UserId:       user.UserId,
			Name:         user.Name,
			Birth:        user.Birth,
			Mobile:       user.Mobile,
			Organization: user.Organization,
			Position:     user.Position,
			Role:         user.Role,
		}
		args["user"] = dao

		return user, nil
	}
	return nil, nil
}

// 반드시 page-initial 초기화 한 후에 실행되어야 함
func InitSidebarArgs(c echo.Context, user *user.User, args map[string]interface{}) error {
	subjects, err := Search_Subjects("", "", "")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, &Result{Error: err})
	}
	if _, has := args["page_initial"]; has {
		args["page_initial"].(map[string]interface{})["sidebar"] = map[string]interface{}{
			"subjects": subjects,
		}
	} else {
		args["page_initial"] = map[string]interface{}{
			"sidebar": map[string]interface{}{
				"subjects": subjects,
			},
		}
	}
	return nil
}

func IsAdmin(c echo.Context) (bool, error) {
	return HasRole(c, "admin")
}

func HasRole(c echo.Context, Role string) (bool, error) {
	RoleName := c.Get(ROLE_KEY)
	if RoleName != nil {
		return (RoleName == Role), nil
	} else {
		return false, nil
	}
}
