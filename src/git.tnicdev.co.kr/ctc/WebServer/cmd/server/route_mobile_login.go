package main

import (
	"net/http"
	"strings"

	"git.tnicdev.co.kr/hami/CTC_Web/pkg/user"

	"github.com/blackss2/utility/convert"
	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo"
)

func route_mobile_login(e *echo.Echo, webChecker echo.MiddlewareFunc) {
	SessionKey := convert.MD5(MOBILE_SECRET_KEYWORD)
	var session_store = sessions.NewCookieStore([]byte(SessionKey + "-store-secret-key"))
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			path := c.Request().URL.Path
			if !strings.HasPrefix(path, "/public") && path != "/favicon.ico" {
				session, err := getSession(c, session_store, SessionKey)
				if err != nil {
					return c.String(http.StatusInternalServerError, err.Error())
				}

				Sid := session.Values[SID_KEY]
				if Sid != nil {
					c.Set(SID_KEY, Sid)
				}
			}

			ret := next(c)
			return ret
		}
	})
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			return next(c) //TEMP

			path := c.Request().URL.Path
			if c.Get(UID_KEY) == nil {
				if !strings.HasPrefix(path, "/public") && path != "/favicon.ico" {
					hash := map[string]bool{
						"/":       true,
						"/login":  true,
						"/logout": true,
					}
					list := []string{
					//"/api/external/",
					}
					hasMatch := false
					if hash[path] {
						hasMatch = true
					} else {
						for _, v := range list {
							if strings.HasPrefix(path, v) {
								hasMatch = true
							}
						}
					}
					if !hasMatch {
						return c.Redirect(http.StatusFound, "/")
					}
				}
			}

			ret := next(c)
			return ret
		}
	})
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			defer context.Clear(c.Request())
			return next(c)
		}
	})

	e.POST("/login", func(c echo.Context) error {
		retCode, retValue := (func() (int, interface{}) {
			scrno := c.FormValue("scrno")
			password := c.FormValue("password")
			//////////////////////////////////////////////////

			session, err := getSession(c, session_store, SessionKey)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			s, err := gAPI.SubjectTable.SubjectByScrNo(scrno)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			if s.Password != password {
				return http.StatusInternalServerError, user.ErrInvalidPassword
			}

			session.Values[SID_KEY] = s.Id

			err = saveSession(c, session)
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
	e.GET("/logout", func(c echo.Context) error {
		if c.Get(SID_KEY) != nil {
			session, _ := getSession(c, session_store, SessionKey)
			session.Values = make(map[interface{}]interface{})
			saveSession(c, session)

			return c.Redirect(http.StatusFound, "/")
		} else {
			return c.Redirect(http.StatusFound, "/")
		}
	}, webChecker)
}
