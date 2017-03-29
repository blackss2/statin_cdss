package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

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

		type Compliance struct {
			Day      string `json:"day"`
			Local    int    `json:"local"`
			Sysmetic int    `json:"sysmetic"`
			Vital    int    `json:"vital"`
			Total    int    `json:"total"`
		}

		if args["IsLogin"] == true {
			total_ae := make(map[string]int)
			arm_ae := make(map[string]int)
			if true {
				subjects, err := gAPI.SubjectTable.List(gConfig.StudyId)
				if err != nil {
					return c.JSON(http.StatusInternalServerError, &Result{Error: err})
				}
				subjectHash := make(map[string]*Subject)
				for _, v := range subjects {
					subjectHash[v.Id] = v
				}

				stacks, err := gAPI.StackTable.ListByFormIds([]string{"f-2", "f-3"}) //TEMP
				if err != nil {
					return c.JSON(http.StatusInternalServerError, &Result{Error: err})
				}

				StackIds := make([]string, 0, len(stacks))
				stackHash := make(map[string]*Stack)
				for _, v := range stacks {
					StackIds = append(StackIds, v.Id)
					stackHash[v.Id] = v
				}
				visits, err := gAPI.VisitTable.ListByStackIds(StackIds)
				if err != nil {
					return c.JSON(http.StatusInternalServerError, &Result{Error: err})
				}

				Ids := make([]string, 0, len(visits))
				visitHash := make(map[string]*Visit)
				for _, v := range visits {
					Ids = append(Ids, v.Id)
					visitHash[v.Id] = v
				}
				dataList, err := gAPI.DataTable.ListByVisitIds(Ids)
				if err != nil {
					return c.JSON(http.StatusInternalServerError, &Result{Error: err})
				}

				subjectAEHash := make(map[string]bool)
				armAEHash := make(map[string]map[string]bool)
				for _, d := range dataList {
					if len(d.Value) > 0 || len(d.CodeId) > 0 {
						if visit, has := visitHash[d.VisitId]; has {
							if stack, has := stackHash[visit.StackId]; has {
								subjectAEHash[stack.SubjectId] = true

								if subject, has := subjectHash[stack.SubjectId]; has {
									hash, has := armAEHash[subject.Arm]
									if !has {
										hash = make(map[string]bool)
										armAEHash[subject.Arm] = hash
									}
									hash[fmt.Sprintf("%s_%d", subject.Id, d.Rowindex)] = true
								}
							}
						}
					}
				}

				total_ae["정상"] = len(subjects) - len(subjectAEHash)
				total_ae["이상반응"] = len(subjectAEHash)

				arm_ae["실험군"] = len(armAEHash["실험군"])
				arm_ae["대조군"] = len(armAEHash["대조군"])
				arm_ae["위약군"] = len(armAEHash["위약군"])
			}
			//TODO
			compliances := make([]*Compliance, 0)
			if true {
				stacks, err := gAPI.StackTable.ListByFormId("f-1") //TEMP
				if err != nil {
					return c.JSON(http.StatusInternalServerError, &Result{Error: err})
				}

				forms, err := FormWithCache(gAPI.FormTable, gConfig.StudyId)
				if err != nil {
					return c.JSON(http.StatusInternalServerError, &Result{Error: err})
				}

				groupHash := make(map[string]*Group)
				itemHash := make(map[string]*Item)
				for _, form := range forms {
					if form.Id == "f-1" {
						addFormMeta(form, groupHash, itemHash)
					}
				}

				countHash := make(map[string]map[string]int)
				StackIds := make([]string, 0, len(stacks))
				for _, v := range stacks {
					StackIds = append(StackIds, v.Id)
				}

				visits, err := gAPI.VisitTable.ListByStackIds(StackIds)
				if err != nil {
					return c.JSON(http.StatusInternalServerError, &Result{Error: err})
				}

				Ids := make([]string, 0, len(visits))
				visitHash := make(map[string]*Visit)
				for _, v := range visits {
					Ids = append(Ids, v.Id)
					visitHash[v.Id] = v
				}
				if len(Ids) > 0 {
					dataList, err := gAPI.DataTable.ListByVisitIds(Ids)
					if err != nil {
						return c.JSON(http.StatusInternalServerError, &Result{Error: err})
					}
					smokerHash := make(map[string]bool)
					alcoholHash := make(map[string]bool)
					for _, d := range dataList {
						if d.ItemId == "i-29" && d.Value != "" {
							if visit, has := visitHash[d.VisitId]; has {
								smokerHash[visit.Position] = true
							}
						}
						if d.ItemId == "i-31" && d.Value != "" {
							if visit, has := visitHash[d.VisitId]; has {
								alcoholHash[visit.Position] = true
							}
						}
					}
					for _, d := range dataList {
						if item, has := itemHash[d.ItemId]; has {
							if visit, has := visitHash[d.VisitId]; has {
								hash, has := countHash[item.GroupId]
								if !has {
									hash = make(map[string]int)
									countHash[item.GroupId] = hash
								}
								hash[visit.Position]++
								if !smokerHash[visit.Position] && d.ItemId == "i-28" && d.CodeId == "c-45" {
									hash[visit.Position]++
								}
								if !alcoholHash[visit.Position] && d.ItemId == "i-30" && d.CodeId == "c-47" {
									hash[visit.Position]++
								}
							}
						}
					}
				}

				DAY_COUNT := 7

				subjectCount, err := gAPI.SubjectTable.Count(gConfig.StudyId)
				if err != nil {
					return c.JSON(http.StatusInternalServerError, &Result{Error: err})
				}

				cpHash := make(map[string]*Compliance)
				for i := 0; i < DAY_COUNT; i++ {
					p := fmt.Sprintf("%d", i+1)
					cp := &Compliance{
						Day:      p,
						Local:    0,
						Sysmetic: 0,
						Vital:    0,
					}
					cpHash[p] = cp
					compliances = append(compliances, cp)
				}
				baseHash := make(map[string]int, DAY_COUNT)
				sumHash := make(map[string]int, DAY_COUNT)
				for i := 0; i < DAY_COUNT; i++ {
					p := fmt.Sprintf("%d", i+1)
					baseHash[p] += len(itemHash) * int(subjectCount)
				}
				for groupid, hash := range countHash {
					itemCount := 0
					for _, v := range itemHash {
						if v.GroupId == groupid {
							itemCount++
						}
					}
					for k, v := range hash {
						switch groupid {
						case "g-2":
							cpHash[k].Local = v * 100 / itemCount / int(subjectCount)
						case "g-3":
							cpHash[k].Sysmetic = v * 100 / itemCount / int(subjectCount)
						case "g-7":
							cpHash[k].Vital = v * 100 / itemCount / int(subjectCount)
						}
						sumHash[k] += v
					}
				}
				for i := 0; i < DAY_COUNT; i++ {
					p := fmt.Sprintf("%d", i+1)
					cpHash[p].Total = sumHash[p] * 100 / baseHash[p]
				}
			}

			//TODO
			enrollment := make(map[string]int)
			//TODO
			reservation := make(map[string]int)
			//TODO

			notices, notice_count, err := Search_Notices(0, 5)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, &Result{Error: err})
			}

			args["page_initial"] = map[string]interface{}{
				"total_ae":     total_ae,
				"arm_ae":       arm_ae,
				"compliances":  compliances,
				"enrollment":   enrollment,
				"reservation":  reservation,
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
	subjects, err := Search_Subjects("", "")
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
