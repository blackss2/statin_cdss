package main

import (
	"bytes"
	"log"
	"net/http"
	"strings"

	"git.tnicdev.co.kr/research/statin_cdss/pkg/user"
	"git.tnicdev.co.kr/research/statin_cdss/pkg/util"

	"github.com/labstack/echo"
)

func main() {
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
	route_api_dupcheck(g)
	route_api_subjects(g)

	//Web Pages
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
	web := util.NewWebServer(e, "./webfiles", StaticFiles)

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
		user, err := InitUserArgs(c, args)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, &Result{Error: err})
		}

		if args["IsLogin"] == true {
			args["page_initial"] = map[string]interface{}{}

			InitSidebarArgs(c, user, args)
			args["has_sidebar"] = true

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
	UserId string `json:"userid"`
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

		var dao = &DAO_Common_UserInfo{
			UserId: user.UserId,
		}
		args["user"] = dao

		return user, nil
	}
	return nil, nil
}

// 반드시 page-initial 초기화 한 후에 실행되어야 함
func InitSidebarArgs(c echo.Context, user *user.User, args map[string]interface{}) error {
	subjects, err := Search_Subjects(user.Id, "", "", "", "")
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
