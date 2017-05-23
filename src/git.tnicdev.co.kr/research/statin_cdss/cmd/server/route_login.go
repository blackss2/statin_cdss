package main

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"git.tnicdev.co.kr/research/statin_cdss/pkg/user"

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
					c.Set(USERID_KEY, session.Values[USERID_KEY])
				} else {
					hash := map[string]bool{
						"/":       true,
						"/login":  true,
						"/logout": true,
						"/join":   true,
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

			u, err := gAPI.UserStore.GetByUserId(userid)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			if u.Password != password {
				return http.StatusInternalServerError, user.ErrInvalidPassword
			}

			session.Values[UID_KEY] = u.Id
			session.Values[USERID_KEY] = u.UserId

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
	e.POST("/join", func(c echo.Context) error {
		retCode, retValue := (func() (int, interface{}) {
			userid := c.FormValue("userid")
			if len(userid) == 0 {
				return http.StatusOK, errors.New("require parameter missing : userid")
			}
			password := c.FormValue("password")
			if len(password) == 0 {
				return http.StatusOK, errors.New("require parameter missing : password")
			}
			password_confirm := c.FormValue("password_confirm")
			if len(password_confirm) == 0 {
				return http.StatusOK, errors.New("require parameter missing : password_confirm")
			}
			if password != password_confirm {
				return http.StatusOK, errors.New("parameter validation violation : password != password_confirm")
			}
			//////////////////////////////////////////////////

			t_create := time.Now()
			_, err := gAPI.UserStore.Insert(userid, password, t_create)
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
