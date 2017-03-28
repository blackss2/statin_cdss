package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"git.tnicdev.co.kr/hami/CTC_Web/pkg/user"
	"git.tnicdev.co.kr/hami/CTC_Web/pkg/util"

	"github.com/PuerkitoBio/goquery"
	"github.com/blackss2/utility/convert"
	"github.com/blackss2/utility/htmlwriter"
	"github.com/labstack/echo"
	"github.com/varstr/uaparser"
)

func main_mobile(wg sync.WaitGroup) {
	defer wg.Done()

	PORT := gConfig.Port.Mobile

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

	//Web Pages
	web := util.NewWebServer(e, "./mobiles")
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
	web.SetupStatic(e, "/public", "./mobiles/public")

	webChecker := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			web.CheckWatch()
			return next(c)
		}
	}
	route_login(e, webChecker)

	e.GET("/", func(c echo.Context) error {
		args := make(map[string]interface{})
		return c.Render(http.StatusOK, "index.html", args)
	})
	e.GET("/main", func(c echo.Context) error {
		TNow := time.Now()

		args := make(map[string]interface{})
		_, err := InitMobileArgs(c, args)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, &Result{Error: err})
		}
		subject, err := gAPI.SubjectTable.Subject("1d74db04-4021-4744-a537-a28e0fdcc0b4") //TEMP
		if err != nil {
			return c.JSON(http.StatusInternalServerError, &Result{Error: err})
		}

		forms, err := FormWithCache(gAPI.FormTable, gConfig.StudyId)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, &Result{Error: err})
		}

		stacks, err := gAPI.StackTable.List(subject.Id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, &Result{Error: err})
		}

		stackByFormId := make(map[string]*Stack)
		for _, v := range stacks {
			stackByFormId[v.FormId] = v
		}

		for _, f := range forms {
			stack := stackByFormId[f.Id]
			visitByPosition := make(map[string]*Visit)
			if stack != nil {
				visits, err := gAPI.VisitTable.List(stack.Id)
				if err != nil {
					return c.JSON(http.StatusInternalServerError, &Result{Error: err})
				}

				for _, v := range visits {
					visitByPosition[v.Position] = v
				}
			}

			if f.Type == "visit" {
				if length, has := f.Extra["visit.length"]; !has {
					return c.JSON(http.StatusInternalServerError, &Result{Error: fmt.Errorf("visit length is not exist")})
				} else {
					list := make([]map[string]interface{}, 0)
					ilen := convert.Int(length)
					firstDate := convert.Time(subject.FirstDate)
					for i := int64(0); i < ilen; i++ {
						hash := make(map[string]interface{})
						targetDate := firstDate.Add(time.Hour * time.Duration(24*i))
						if i == 0 {
							hash["Name"] = "접종일"
						} else {
							hash["Name"] = fmt.Sprintf("접종 후 %d일", i)
						}
						hash["Date"] = convert.String(targetDate)[:10]
						if !targetDate.After(TNow) {
							hash["Active"] = true
						}
						position := (i + 1)
						hash["Position"] = position
						if _, has := visitByPosition[convert.String(position)]; has {
							hash["Saved"] = true
						}
						list = append(list, hash)
					}
					args["VisitList"] = list
				}
			}
		}
		args["Forms"] = forms
		return c.Render(http.StatusOK, "main.html", args)
	})
	e.GET("/list", func(c echo.Context) error {
		args := make(map[string]interface{})
		user, err := InitMobileArgs(c, args)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, &Result{Error: err})
		}

		TNow := time.Now()
		ActorId := user.Id

		subject, err := gAPI.SubjectTable.Subject("1d74db04-4021-4744-a537-a28e0fdcc0b4") //TEMP
		if err != nil {
			return c.JSON(http.StatusInternalServerError, &Result{Error: err})
		}

		id := c.QueryParam("id")
		gid := c.QueryParam("gid")

		forms, err := FormWithCache(gAPI.FormTable, gConfig.StudyId)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, &Result{Error: err})
		}

		var form *Form
		for _, f := range forms {
			if f.Id == id {
				form = f
				break
			}
		}
		if form == nil {
			return c.NoContent(http.StatusNotFound)
		}

		stack, err := gAPI.StackTable.Stack(subject.Id, form.Id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, &Result{Error: err})
		}
		if stack == nil {
			s, err := gAPI.StackTable.Insert(subject.Id, form.Id, TNow, ActorId)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, &Result{Error: err})
			}
			stack = s
		}

		position := "1"
		visit, err := gAPI.VisitTable.Visit(stack.Id, position)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, &Result{Error: err})
		}
		if visit == nil {
			v, err := gAPI.VisitTable.Insert(stack.Id, position, TNow, ActorId)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, &Result{Error: err})
			}
			visit = v
		}

		dataList, err := gAPI.DataTable.List(visit.Id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, &Result{Error: err})
		}

		dataHash := make(map[int64][]*Data)
		maxRowindex := int64(0)
		for _, data := range dataList {
			if maxRowindex < data.Rowindex {
				maxRowindex = data.Rowindex
			}
			list, has := dataHash[data.Rowindex]
			if !has {
				list = make([]*Data, 0)
			}
			dataHash[data.Rowindex] = append(list, data)
		}
		dataRowindexList := make([][]*Data, maxRowindex+1)
		for rowindex, list := range dataHash {
			dataRowindexList[rowindex] = list
		}

		args["Form"] = form
		args["NextRowindex"] = len(dataRowindexList)

		groupHash := make(map[string]*Group)
		itemHash := make(map[string]*Item)
		for _, form := range forms {
			addFormMeta(form, groupHash, itemHash)
		}

		var group *Group
		for _, g := range form.Groups {
			if g.Id == gid {
				group = g
				break
			}
		}

		if group == nil {
			return c.NoContent(http.StatusNotFound)
		}

		args["Title"] = group.Name

		jRoot := htmlwriter.CreateHtmlNode("div").Class("form-grp")
		jRoot.Attr("formid", form.Id)
		jRoot.Attr("position", position)
		err = group.GenerateHTML(position, jRoot, true)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, &Result{Error: err})
		}

		var buffer bytes.Buffer
		jRoot.WriteWith(&buffer, "\t")

		var htmlBuffer bytes.Buffer
		if true {
			docOrg, err := goquery.NewDocumentFromReader(&buffer)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, &Result{Error: err})
			}

			for r, list := range dataRowindexList {
				if list == nil {
					continue
				}

				//apply data to html
				doc := docOrg.Clone()
				jFormGrp := doc.Find(".form-grp")

				jFormGrp.SetAttr("rowindex", convert.String(r))
				//TODO : rowindex based iteration
				for _, data := range list {
					if len(data.Value) > 0 || len(data.CodeId) > 0 {
						item, has := itemHash[data.ItemId]
						if !has {
							return c.JSON(http.StatusInternalServerError, &Result{Error: fmt.Errorf("no item : %v", data.ItemId)})
						}

						formKey := data.ItemId
						jTarget := jFormGrp.Find(fmt.Sprintf("[name='%s']", formKey))
						if jTarget.Length() == 0 {
							return c.JSON(http.StatusInternalServerError, &Result{Error: fmt.Errorf("target length is zero")})
						}

						var Value string
						switch item.Type {
						case "checkbox":
							fallthrough
						case "radio":
							Value = data.CodeId
						default:
							Value = data.Value
						}

						switch item.Type {
						case "checkbox":
							fallthrough
						case "radio":
							jTarget.Filter(fmt.Sprintf(":not([value='%s'])", Value)).Parent().Parent().Remove()
							jTarget = jTarget.Filter(fmt.Sprintf("[value='%s']", Value))
							text := strings.TrimSpace(jTarget.Next().Find(".item-title").Text())
							jItemInner := jTarget.Parent().Parent().Parent().Parent()
							jItemInner.Empty()
							jItemInner.AppendHtml(strings.Join(strings.Split(text, "\n"), "<br>"))
						default:
							jTarget.SetAttr("value", Value)
						}
					}
				}

				jFormGrp.Find("[itemtype]").Each(func(idx int, jObj *goquery.Selection) {
					itemtype, _ := jObj.Attr("itemtype")
					text, _ := jObj.Attr("value")
					switch itemtype {
					case "checkbox":
						fallthrough
					case "radio":
					default:
						jParnet := jObj.Parent()
						jParnet.Empty()
						jParnet.AppendHtml(strings.Join(strings.Split(text, "\n"), "<br>"))
					}
				})

				ret, err := doc.Html()
				if err != nil {
					return c.JSON(http.StatusInternalServerError, &Result{Error: err})
				}

				htmlBuffer.WriteString(ret)
			}
		}

		html := htmlBuffer.String()
		args["FormHtml"] = template.HTML(html)

		return c.Render(http.StatusOK, "list.html", args)
	})
	e.GET("/form", func(c echo.Context) error {
		args := make(map[string]interface{})
		user, err := InitMobileArgs(c, args)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, &Result{Error: err})
		}

		TNow := time.Now()
		ActorId := user.Id

		subject, err := gAPI.SubjectTable.Subject("1d74db04-4021-4744-a537-a28e0fdcc0b4") //TEMP
		if err != nil {
			return c.JSON(http.StatusInternalServerError, &Result{Error: err})
		}

		id := c.QueryParam("id")
		position := c.QueryParam("position")
		if len(position) == 0 {
			position = "1"
		}
		rowindex := convert.Int(c.QueryParam("rowindex"))

		forms, err := FormWithCache(gAPI.FormTable, gConfig.StudyId)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, &Result{Error: err})
		}

		var form *Form
		for _, f := range forms {
			if f.Id == id {
				form = f
				break
			}
		}
		if form == nil {
			return c.NoContent(http.StatusNotFound)
		}

		stack, err := gAPI.StackTable.Stack(subject.Id, form.Id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, &Result{Error: err})
		}
		if stack == nil {
			s, err := gAPI.StackTable.Insert(subject.Id, form.Id, TNow, ActorId)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, &Result{Error: err})
			}
			stack = s
		}

		visit, err := gAPI.VisitTable.Visit(stack.Id, position)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, &Result{Error: err})
		}
		if visit == nil {
			v, err := gAPI.VisitTable.Insert(stack.Id, position, TNow, ActorId)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, &Result{Error: err})
			}
			visit = v
		}

		dataList, err := gAPI.DataTable.ListByRowindex(visit.Id, rowindex)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, &Result{Error: err})
		}

		if form.Type == "visit" {
			args["IsNew"] = true
		} else if len(dataList) == 0 {
			args["IsNew"] = true
		}

		groupHash := make(map[string]*Group)
		itemHash := make(map[string]*Item)
		for _, form := range forms {
			addFormMeta(form, groupHash, itemHash)
		}

		jRoot := htmlwriter.CreateHtmlNode("div").Class("form-grp")
		jRoot.Attr("formid", form.Id)
		jRoot.Attr("position", position)
		jRoot.Attr("rowindex", convert.String(rowindex))
		err = form.GenerateHTML(position, jRoot)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, &Result{Error: err})
		}

		var buffer bytes.Buffer
		jRoot.WriteWith(&buffer, "\t")

		html := buffer.String()
		if true {
			doc, err := goquery.NewDocumentFromReader(&buffer)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, &Result{Error: err})
			}

			//apply data to html
			jFormGrp := doc.Find(".form-grp")
			for _, data := range dataList {
				if len(data.Value) > 0 || len(data.CodeId) > 0 {
					item, has := itemHash[data.ItemId]
					if !has {
						return c.JSON(http.StatusInternalServerError, &Result{Error: fmt.Errorf("no item : %v", data.ItemId)})
					}

					formKey := data.ItemId
					jTarget := jFormGrp.Find(fmt.Sprintf("[name='%s']", formKey))
					if jTarget.Length() == 0 {
						return c.JSON(http.StatusInternalServerError, &Result{Error: fmt.Errorf("target length is zero")})
					}

					var Value string
					switch item.Type {
					case "checkbox":
						fallthrough
					case "radio":
						Value = data.CodeId
					default:
						Value = data.Value
					}

					switch item.Type {
					case "checkbox":
						fallthrough
					case "radio":
						jTarget.Filter(fmt.Sprintf("[value='%s']", Value)).SetAttr("checked", "true")
					case "textarea":
						jTarget.AppendHtml(Value)
					default:
						jTarget.SetAttr("value", Value)
					}
				}
			}

			firstDate := convert.Time(subject.FirstDate)
			doc.Find("[name='i-1']").SetAttr("value", convert.String(firstDate.Add(time.Hour * time.Duration(24*(convert.Int(position)-1))))[:10])

			ret, err := doc.Html()
			if err != nil {
				return c.JSON(http.StatusInternalServerError, &Result{Error: err})
			}

			html = ret
		}

		args["FormHtml"] = template.HTML(html)

		return c.Render(http.StatusOK, "form.html", args)
	})
	e.POST("/form", func(c echo.Context) error {
		args := make(map[string]interface{})
		user, err := InitMobileArgs(c, args)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, &Result{Error: err})
		}

		TNow := time.Now()
		ActorId := user.Id

		subject, err := gAPI.SubjectTable.Subject("1d74db04-4021-4744-a537-a28e0fdcc0b4") //TEMP
		if err != nil {
			return c.JSON(http.StatusInternalServerError, &Result{Error: err})
		}

		id := c.QueryParam("id")
		position := c.QueryParam("position")
		if len(position) == 0 {
			position = "1"
		}
		rowindex := convert.Int(c.QueryParam("rowindex"))

		forms, err := FormWithCache(gAPI.FormTable, gConfig.StudyId)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, &Result{Error: err})
		}

		var form *Form
		for _, f := range forms {
			if f.Id == id {
				form = f
				break
			}
		}
		if form == nil {
			return c.NoContent(http.StatusNotFound)
		}

		stack, err := gAPI.StackTable.Stack(subject.Id, form.Id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, &Result{Error: err})
		}
		if stack == nil {
			s, err := gAPI.StackTable.Insert(subject.Id, form.Id, TNow, ActorId)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, &Result{Error: err})
			}
			stack = s
		}

		visit, err := gAPI.VisitTable.Visit(stack.Id, position)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, &Result{Error: err})
		}
		if visit == nil {
			v, err := gAPI.VisitTable.Insert(stack.Id, position, TNow, ActorId)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, &Result{Error: err})
			}
			visit = v
		}

		dataList, err := gAPI.DataTable.ListByRowindex(visit.Id, rowindex)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, &Result{Error: err})
		}

		if form.Type == "visit" {
			args["IsNew"] = true
		} else if len(dataList) == 0 {
			args["IsNew"] = true
		}

		groupHash := make(map[string]*Group)
		itemHash := make(map[string]*Item)
		for _, form := range forms {
			addFormMeta(form, groupHash, itemHash)
		}

		dataHash := make(map[string]*Data)
		for _, data := range dataList {
			item, has := itemHash[data.ItemId]
			if !has {
				return c.JSON(http.StatusInternalServerError, &Result{Error: fmt.Errorf("no item : %v", data)})
			}

			var dataKey string
			switch item.Type {
			case "checkbox":
				dataKey = fmt.Sprintf("%s_%d_%s", data.ItemId, data.Rowindex, data.CodeId)
			default:
				dataKey = fmt.Sprintf("%s_%d", data.ItemId, data.Rowindex)
			}
			dataHash[dataKey] = data
		}

		SaveData, _ := c.FormParams()
		insertList := make([]*Data, 0)
		historyList := make([]*History, 0)
		for ItemId, formValueList := range SaveData {
			if !strings.HasPrefix(ItemId, "i-") {
				continue
			}

			item, has := itemHash[ItemId]
			if !has {
				return c.JSON(http.StatusInternalServerError, &Result{Error: fmt.Errorf("no item : %s", ItemId)})
			}

			for _, formValue := range formValueList {
				var Value string
				var CodeId string
				switch item.Type {
				case "radio":
					fallthrough
				case "checkbox":
					CodeId = formValue
				default:
					Value = formValue
				}

				var dataKey string
				switch item.Type {
				case "checkbox":
					dataKey = fmt.Sprintf("%s_%d_%s", item.Id, rowindex, CodeId)
				default:
					dataKey = fmt.Sprintf("%s_%d", item.Id, rowindex)
				}

				if len(Value) > 0 || len(CodeId) > 0 {
					if data, has := dataHash[dataKey]; has {
						if data.Value != Value || data.CodeId != CodeId {
							//send current data to history
							history := gAPI.HistoryTable.Retain(data, TNow, ActorId)
							historyList = append(historyList, history)

							//update data
							data.Value = Value
							data.CodeId = CodeId
							data.ActorId = ActorId
							data.TCreate = TNow
							err := gAPI.DataTable.Update(data.Id, data)
							if err != nil {
								return c.JSON(http.StatusInternalServerError, &Result{Error: err})
							}
						}
						delete(dataHash, dataKey)
					} else {
						data = &Data{
							Value:    Value,
							CodeId:   CodeId,
							Rowindex: rowindex,
							ItemId:   ItemId,
							VisitId:  visit.Id,
							TCreate:  TNow,
							ActorId:  ActorId,
						}
						insertList = append(insertList, data)
					}
				}
			}
		}

		DeleteIdList := make([]string, 0)
		for _, data := range dataHash {
			//send current data to history
			history := gAPI.HistoryTable.Retain(data, TNow, ActorId)
			historyList = append(historyList, history)

			DeleteIdList = append(DeleteIdList, data.Id)
		}
		if len(DeleteIdList) > 0 {
			err := gAPI.DataTable.DeleteById("visit_id", visit.Id, DeleteIdList)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, &Result{Error: err})
			}
		}

		if len(insertList) > 0 {
			//insert data
			err := gAPI.DataTable.Insert(insertList)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, &Result{Error: err})
			}
		}

		if len(historyList) > 0 {
			//send current data to history
			err := gAPI.HistoryTable.Insert(historyList)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, &Result{Error: err})
			}
		}
		return c.NoContent(http.StatusOK)
	})
	e.DELETE("/form", func(c echo.Context) error {
		args := make(map[string]interface{})
		user, err := InitMobileArgs(c, args)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, &Result{Error: err})
		}

		TNow := time.Now()
		ActorId := user.Id

		subject, err := gAPI.SubjectTable.Subject("1d74db04-4021-4744-a537-a28e0fdcc0b4") //TEMP
		if err != nil {
			return c.JSON(http.StatusInternalServerError, &Result{Error: err})
		}

		id := c.QueryParam("id")
		position := c.QueryParam("position")
		if len(position) == 0 {
			position = "1"
		}
		rowindex := convert.Int(c.QueryParam("rowindex"))

		forms, err := FormWithCache(gAPI.FormTable, gConfig.StudyId)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, &Result{Error: err})
		}

		var form *Form
		for _, f := range forms {
			if f.Id == id {
				form = f
				break
			}
		}
		if form == nil {
			return c.NoContent(http.StatusNotFound)
		}

		stack, err := gAPI.StackTable.Stack(subject.Id, form.Id)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, &Result{Error: err})
		}
		if stack == nil {
			s, err := gAPI.StackTable.Insert(subject.Id, form.Id, TNow, ActorId)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, &Result{Error: err})
			}
			stack = s
		}

		visit, err := gAPI.VisitTable.Visit(stack.Id, position)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, &Result{Error: err})
		}
		if visit == nil {
			v, err := gAPI.VisitTable.Insert(stack.Id, position, TNow, ActorId)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, &Result{Error: err})
			}
			visit = v
		}

		dataList, err := gAPI.DataTable.ListByRowindex(visit.Id, rowindex)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, &Result{Error: err})
		}

		if form.Type == "visit" {
			args["IsNew"] = true
		} else if len(dataList) == 0 {
			args["IsNew"] = true
		}

		groupHash := make(map[string]*Group)
		itemHash := make(map[string]*Item)
		for _, form := range forms {
			addFormMeta(form, groupHash, itemHash)
		}

		if form.Type != "visit" {
			DeleteIdList := make([]string, 0)

			historyList := make([]*History, 0)
			for _, data := range dataList {
				history := gAPI.HistoryTable.Retain(data, TNow, ActorId)
				historyList = append(historyList, history)
				DeleteIdList = append(DeleteIdList, data.Id)
			}
			if len(DeleteIdList) > 0 {
				err := gAPI.DataTable.DeleteById("visit_id", visit.Id, DeleteIdList)
				if err != nil {
					return c.JSON(http.StatusInternalServerError, &Result{Error: err})
				}
			}
			if len(historyList) > 0 {
				//send current data to history
				err := gAPI.HistoryTable.Insert(historyList)
				if err != nil {
					return c.JSON(http.StatusInternalServerError, &Result{Error: err})
				}
			}
		}
		return c.NoContent(http.StatusOK)
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

func InitMobileArgs(c echo.Context, args map[string]interface{}) (*user.User, error) {
	ua := c.Request().Header.Get("User-Agent")
	rs := uaparser.Parse(ua)
	log.Println(rs.Browser, rs.Device, rs.DeviceType, rs.OS)

	styleHash := make(map[string]interface{})
	styleHash["icon_only"] = ""
	styleHash["pages_style"] = "navbar-fixed"
	args["Style"] = styleHash

	return InitUserArgs(c, args)
}
