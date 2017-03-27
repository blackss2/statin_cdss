package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"git.tnicdev.co.kr/hami/CTC_Web/pkg/util"

	"github.com/PuerkitoBio/goquery"
	"github.com/blackss2/utility/convert"
	"github.com/blackss2/utility/htmlwriter"
	"github.com/gin-gonic/gin"
	"github.com/varstr/uaparser"
)

func main_mobile(wg sync.WaitGroup) {
	defer wg.Done()

	PORT := gConfig.Port.Mobile

	WebPath, err := filepath.Abs("./mobiles")
	if err != nil {
		log.Fatalln(err)
	}

	isRequireReload := true
	util.NewFileWatcher(WebPath, func(ev string, path string) {
		if strings.HasPrefix(filepath.Ext(path), ".htm") {
			isRequireReload = true
		}
	})

	r := gin.Default()
	r.GET("/public/*path", func(c *gin.Context) {
		uri := c.Request.URL.Path
		c.File(WebPath + uri)
	})

	templateHash := make(map[string]*template.Template)
	r.NoRoute(func(c *gin.Context) {
		TNow := time.Now()
		ActorId := "user"

		path := c.Request.URL.Path

		if isRequireReload {
			hash, err := loadTemplate(WebPath)
			if err != nil {
				c.AbortWithError(500, err)
				return
			}
			templateHash = hash
			isRequireReload = false
		}

		if tp, has := templateHash[path]; has {
			ua := c.Request.Header.Get("User-Agent")
			rs := uaparser.Parse(ua)
			log.Println(rs.Browser, rs.Device, rs.DeviceType, rs.OS)

			args := make(map[string]interface{})
			styleHash := make(map[string]interface{})
			styleHash["icon_only"] = ""
			styleHash["pages_style"] = "navbar-fixed"
			args["Style"] = styleHash

			var subject *Subject
			if path != "/" {
				s, err := gAPI.SubjectTable.Subject("1d74db04-4021-4744-a537-a28e0fdcc0b4") //TEMP
				if err != nil {
					c.AbortWithError(500, err)
				}
				subject = s

				if len(s.FirstDate) == 0 {
					s.FirstDate = "2016-03-26" //TEMP
				}
			}

			switch path {
			case "/main":
				forms, err := FormWithCache(gAPI.FormTable, gConfig.StudyId)
				if err != nil {
					c.AbortWithError(500, err)
					return
				}

				stacks, err := gAPI.StackTable.List(subject.Id)
				if err != nil {
					c.AbortWithError(500, err)
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
							c.AbortWithError(500, err)
						}

						for _, v := range visits {
							visitByPosition[v.Position] = v
						}
					}

					if f.Type == "visit" {
						if length, has := f.Extra["visit.length"]; !has {
							c.AbortWithError(500, fmt.Errorf("visit length is not exist"))
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
			case "/list":
				id := c.Query("id")
				gid := c.Query("gid")

				forms, err := FormWithCache(gAPI.FormTable, gConfig.StudyId)
				if err != nil {
					c.AbortWithError(500, err)
					return
				}

				var form *Form
				for _, f := range forms {
					if f.Id == id {
						form = f
						break
					}
				}
				if form == nil {
					c.Status(404)
					return
				}

				stack, err := gAPI.StackTable.Stack(subject.Id, form.Id)
				if err != nil {
					c.AbortWithError(500, err)
				}
				if stack == nil {
					s, err := gAPI.StackTable.Insert(subject.Id, form.Id, TNow, ActorId)
					if err != nil {
						c.AbortWithError(500, err)
					}
					stack = s
				}

				position := "1"
				visit, err := gAPI.VisitTable.Visit(stack.Id, position)
				if err != nil {
					c.AbortWithError(500, err)
				}
				if visit == nil {
					v, err := gAPI.VisitTable.Insert(stack.Id, position, TNow, ActorId)
					if err != nil {
						c.AbortWithError(500, err)
					}
					visit = v
				}

				dataList, err := gAPI.DataTable.List(visit.Id)
				if err != nil {
					c.AbortWithError(500, err)
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
					c.Status(404)
					return
				}

				args["Title"] = group.Name

				jRoot := htmlwriter.CreateHtmlNode("div").Class("form-grp")
				jRoot.Attr("formid", form.Id)
				jRoot.Attr("position", position)
				err = group.GenerateHTML(position, jRoot, true)
				if err != nil {
					c.AbortWithError(500, err)
					return
				}

				var buffer bytes.Buffer
				jRoot.WriteWith(&buffer, "\t")

				var htmlBuffer bytes.Buffer
				if true {
					docOrg, err := goquery.NewDocumentFromReader(&buffer)
					if err != nil {
						c.AbortWithError(500, err)
						return
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
									c.AbortWithError(500, fmt.Errorf("no item : %v", data.ItemId))
									return
								}

								formKey := data.ItemId
								jTarget := jFormGrp.Find(fmt.Sprintf("[name='%s']", formKey))
								if jTarget.Length() == 0 {
									c.AbortWithError(500, fmt.Errorf("target length is zero"))
									return
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
							c.AbortWithError(500, err)
							return
						}

						htmlBuffer.WriteString(ret)
					}
				}

				html := htmlBuffer.String()
				args["FormHtml"] = template.HTML(html)
			case "/form":
				id := c.Query("id")
				position := c.Query("position")
				if len(position) == 0 {
					position = "1"
				}
				rowindex := convert.Int(c.Query("rowindex"))

				forms, err := FormWithCache(gAPI.FormTable, gConfig.StudyId)
				if err != nil {
					c.AbortWithError(500, err)
					return
				}

				var form *Form
				for _, f := range forms {
					if f.Id == id {
						form = f
						break
					}
				}
				if form == nil {
					c.Status(404)
					return
				}

				stack, err := gAPI.StackTable.Stack(subject.Id, form.Id)
				if err != nil {
					c.AbortWithError(500, err)
				}
				if stack == nil {
					s, err := gAPI.StackTable.Insert(subject.Id, form.Id, TNow, ActorId)
					if err != nil {
						c.AbortWithError(500, err)
					}
					stack = s
				}

				visit, err := gAPI.VisitTable.Visit(stack.Id, position)
				if err != nil {
					c.AbortWithError(500, err)
				}
				if visit == nil {
					v, err := gAPI.VisitTable.Insert(stack.Id, position, TNow, ActorId)
					if err != nil {
						c.AbortWithError(500, err)
					}
					visit = v
				}

				dataList, err := gAPI.DataTable.ListByRowindex(visit.Id, rowindex)
				if err != nil {
					c.AbortWithError(500, err)
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

				switch c.Request.Method {
				case "DELETE":
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
								c.AbortWithError(500, err)
								return
							}
						}
						if len(historyList) > 0 {
							//send current data to history
							err := gAPI.HistoryTable.Insert(historyList)
							if err != nil {
								c.AbortWithError(500, err)
								return
							}
						}
					}
				case "GET":
					jRoot := htmlwriter.CreateHtmlNode("div").Class("form-grp")
					jRoot.Attr("formid", form.Id)
					jRoot.Attr("position", position)
					jRoot.Attr("rowindex", convert.String(rowindex))
					err = form.GenerateHTML(position, jRoot)
					if err != nil {
						c.AbortWithError(500, err)
						return
					}

					var buffer bytes.Buffer
					jRoot.WriteWith(&buffer, "\t")

					html := buffer.String()
					if true {
						doc, err := goquery.NewDocumentFromReader(&buffer)
						if err != nil {
							c.AbortWithError(500, err)
							return
						}

						//apply data to html
						jFormGrp := doc.Find(".form-grp")
						for _, data := range dataList {
							if len(data.Value) > 0 || len(data.CodeId) > 0 {
								item, has := itemHash[data.ItemId]
								if !has {
									c.AbortWithError(500, fmt.Errorf("no item : %v", data.ItemId))
									return
								}

								formKey := data.ItemId
								jTarget := jFormGrp.Find(fmt.Sprintf("[name='%s']", formKey))
								if jTarget.Length() == 0 {
									c.AbortWithError(500, fmt.Errorf("target length is zero"))
									return
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
							c.AbortWithError(500, err)
							return
						}

						html = ret
					}

					args["FormHtml"] = template.HTML(html)
				case "POST":
					err := c.Request.ParseForm()
					if err != nil {
						c.AbortWithError(500, err)
						return
					}

					dataHash := make(map[string]*Data)
					for _, data := range dataList {
						item, has := itemHash[data.ItemId]
						if !has {
							c.AbortWithError(500, fmt.Errorf("no item : %v", data))
							return
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

					SaveData := c.Request.Form
					insertList := make([]*Data, 0)
					historyList := make([]*History, 0)
					for ItemId, formValueList := range SaveData {
						if !strings.HasPrefix(ItemId, "i-") {
							continue
						}

						item, has := itemHash[ItemId]
						if !has {
							c.AbortWithError(500, fmt.Errorf("no item : %s", ItemId))
							return
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
											c.AbortWithError(500, err)
											return
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
							c.AbortWithError(500, err)
							return
						}
					}

					if len(insertList) > 0 {
						//insert data
						err := gAPI.DataTable.Insert(insertList)
						if err != nil {
							c.AbortWithError(500, err)
							return
						}
					}

					if len(historyList) > 0 {
						//send current data to history
						err := gAPI.HistoryTable.Insert(historyList)
						if err != nil {
							c.AbortWithError(500, err)
							return
						}
					}

					c.Status(200)
					return
				}
			}

			c.Status(200)
			c.Writer.Header().Set("Content-type", "text/html; charset=UTF-8")
			err := tp.Execute(c.Writer, args)
			if err != nil {
				c.AbortWithError(500, err)
				return
			}
		} else {
			c.String(404, "404 Not Found")
		}
	})

	err = r.Run(":" + PORT)
	if err != nil {
		log.Fatalln(err)
	}
}

func addFormMeta(form *Form, groupHash map[string]*Group, itemHash map[string]*Item) {
	//load meta
	for _, group := range form.Groups {
		groupHash[group.Id] = group
		for _, item := range group.Items {
			itemHash[item.Id] = item
		}
	}
}

func loadTemplate(WebPath string) (map[string]*template.Template, error) {
	templateHash := make(map[string]*template.Template)
	err := filepath.Walk(WebPath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			if strings.HasPrefix(filepath.Ext(path), ".htm") {
				relPath, err := filepath.Rel(WebPath, path)
				if err != nil {
					return err
				}
				data, err := ioutil.ReadFile(path)
				if err != nil {
					return err
				}
				tp, err := template.New("main").Parse(string(data))
				if err != nil {
					return err
				}
				ext := filepath.Ext(relPath)
				name := relPath[:len(relPath)-len(ext)]
				if name == "index" {
					name = ""
				}
				templateHash["/"+name] = tp
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return templateHash, nil
}
