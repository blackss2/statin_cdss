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

func route_login(e *echo.Echo, webChecker echo.MiddlewareFunc) {
	SessionKey := convert.MD5(SECRET_KEYWORD)
	var session_store = sessions.NewCookieStore([]byte(SessionKey + "-store-secret-key"))
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			path := c.Request().URL.Path
			if !strings.HasPrefix(path, "/public") && path != "/favicon.ico" {
				session, err := getSession(c, session_store, SessionKey)
				if err != nil {
					return c.String(http.StatusInternalServerError, err.Error())
				}

				Uid := session.Values[UID_KEY]
				if Uid != nil {
					c.Set(UID_KEY, Uid)
				}
				UserId := session.Values[USERID_KEY]
				if UserId != nil {
					c.Set(USERID_KEY, UserId)
				}
				Role := session.Values[ROLE_KEY]
				if Role != nil {
					c.Set(ROLE_KEY, Role)
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

	e.GET("/login", func(c echo.Context) error {
		return c.Render(http.StatusOK, "login.html", nil)
	}, webChecker)
	e.POST("/login", func(c echo.Context) error {
		retCode, retValue := (func() (int, interface{}) {
			userid := c.FormValue("userid")
			password := c.FormValue("password")
			//////////////////////////////////////////////////

			session, err := getSession(c, session_store, SessionKey)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			u, err := gAPI.UserStore.GetByUserId(userid, true)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			if u.Password != password {
				return http.StatusInternalServerError, user.ErrInvalidPassword
			}

			session.Values[UID_KEY] = u.Id
			session.Values[USERID_KEY] = u.UserId
			session.Values[ROLE_KEY] = u.Role

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
		if c.Get(USERID_KEY) != nil {
			if c.Get(UID_KEY) != nil {
				Uid := c.Get(UID_KEY).(string)

				var user *user.User
				err := gAPI.UserStore.Get(Uid, &user)
				if err != nil {
					return c.JSON(http.StatusInternalServerError, &Result{Error: err})
				}
			}

			session, _ := getSession(c, session_store, SessionKey)
			session.Values = make(map[interface{}]interface{})
			saveSession(c, session)

			return c.Redirect(http.StatusFound, "/")
		} else {
			return c.Redirect(http.StatusFound, "/")
		}
	}, webChecker)
}

func getSession(c echo.Context, session_store *sessions.CookieStore, SessionKey string) (*sessions.Session, error) {
	r := c.Request()
	session, err := session_store.Get(r, SessionKey)
	if err != nil {
		session, err = session_store.New(r, SessionKey)
		if err != nil {
			return nil, err
		}
	}
	return session, nil
}
func saveSession(c echo.Context, session *sessions.Session) error {
	return session.Save(c.Request(), c.Response())
}
