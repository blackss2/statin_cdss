package util

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/labstack/echo"
	"github.com/wellington/go-libsass"
)

type StaticFilesFile struct {
	Data  string
	Mime  string
	Mtime time.Time
	// size is the size before compression. If 0, it means the data is uncompressed
	Size int
	// hash is a sha256 hash of the file contents. Used for the Etag, and useful for caching
	Hash string
}

type WebServer struct {
	path            string
	hasWatch        bool
	templates       *template.Template
	echo            *echo.Echo
	staticFiles     map[string]*StaticFilesFile
	isRequireReload bool
	sync.Mutex
}

func NewWebServer(echo *echo.Echo, path string, sf map[string]*StaticFilesFile) *WebServer {
	web := &WebServer{
		echo:        echo,
		path:        path,
		staticFiles: sf,
	}

	if fi, err := os.Stat(path); err == nil && fi.IsDir() {
		WebPath, err := filepath.Abs(path)
		if err != nil {
			log.Fatalln(err)
		}

		NewFileWatcher(WebPath, func(ev string, path string) {
			if strings.HasPrefix(filepath.Ext(path), ".htm") {
				web.isRequireReload = true
			}
		})
		web.hasWatch = true
	}
	web.UpdateRender()

	return web
}

func (web *WebServer) CheckWatch() {
	if web.isRequireReload {
		web.Lock()
		if web.isRequireReload {
			err := web.UpdateRender()
			if err != nil {
				log.Println(err)
			} else {
				web.isRequireReload = false
			}
		}
		web.Unlock()
	}
}

func (web *WebServer) UpdateRender() error {
	tp := template.New("").Delims("<%", "%>").Funcs(template.FuncMap{
		"marshal": func(v interface{}) string {
			a, _ := json.Marshal(v)
			return string(a)
		},
	})
	if web.hasWatch {
		filepath.Walk(web.path, func(path string, fi os.FileInfo, err error) error {
			if strings.HasPrefix(filepath.Ext(path), ".htm") {
				rel, err := filepath.Rel(web.path, path)
				if err != nil {
					return err
				}
				data, err := ioutil.ReadFile(path)
				if err != nil {
					return err
				}
				rel = filepath.ToSlash(rel)
				template.Must(tp.New(rel).Parse(string(data)))
			}
			return nil
		})
	} else {
		for path, v := range web.staticFiles {
			if strings.HasPrefix(filepath.Ext(path), ".htm") {
				var data []byte
				if v.Size == 0 {
					data = []byte(v.Data)
				} else {
					br := bytes.NewReader([]byte(v.Data))
					r, err := gzip.NewReader(br)
					body, err := ioutil.ReadAll(r)
					if err != nil {
						panic(err)
					}
					data = body
				}
				template.Must(tp.New(path).Parse(string(data)))
			}
		}
	}
	web.templates = tp

	return nil
}

func (web *WebServer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return web.templates.ExecuteTemplate(w, name, data)
}

func (web *WebServer) SetupStatic(e *echo.Echo, prefix string, root string) {
	h := func(c echo.Context) error {
		fname := c.Param("*")
		upath := path.Join(prefix, fname)[1:]
		ext := path.Ext(upath)
		IsCss := strings.HasPrefix(upath, "public/css/") && ext == ".css"
		if IsCss {
			var data []byte
			if web.hasWatch {
				fpath := path.Join(root, fname)
				b, err := ioutil.ReadFile(fpath)
				if err != nil {
					return err
				}
				data = b
			} else {
				if file, has := web.staticFiles[upath]; !has {
					return c.NoContent(http.StatusNotFound)
				} else {
					reader, err := gzip.NewReader(strings.NewReader(file.Data))
					if err != nil {
						return err
					}
					defer reader.Close()
					b, err := ioutil.ReadAll(reader)
					if err != nil {
						return err
					}
					data = b
				}
			}
			if len(data) == 0 {
				return c.NoContent(http.StatusOK)
			} else {
				var buffer bytes.Buffer
				comp, err := libsass.New(&buffer, bytes.NewReader(data))
				if err != nil {
					return err
				}

				err = comp.Run()
				if err != nil {
					return err
				}

				rw := c.Response().Writer
				header := c.Response().Header()
				header.Set("Content-Type", "text/css")
				bs := buffer.Bytes()
				header.Set("Content-Length", strconv.Itoa(len(bs)))
				c.Response().WriteHeader(http.StatusOK)
				_, err = io.Copy(rw, &buffer)
				if err != nil {
					return err
				}
				return nil
			}
		} else {
			if web.hasWatch {
				fpath := path.Join(root, fname)
				return c.File(fpath)
			} else {
				if file, has := web.staticFiles[upath]; !has {
					return c.NoContent(http.StatusNotFound)
				} else {
					rw := c.Response().Writer
					header := c.Response().Header()
					req := c.Request()

					if file.Hash != "" {
						if hash := req.Header.Get("If-None-Match"); hash == file.Hash {
							c.Response().WriteHeader(http.StatusNotModified)
							return nil
						}
						header.Set("ETag", file.Hash)
					}
					if !file.Mtime.IsZero() {
						if t, err := time.Parse(http.TimeFormat, req.Header.Get("If-Modified-Since")); err == nil && file.Mtime.Before(t.Add(1*time.Second)) {
							c.Response().WriteHeader(http.StatusNotModified)
							return nil
						}
						header.Set("Last-Modified", file.Mtime.UTC().Format(http.TimeFormat))
					}

					header.Set("Content-Type", file.Mime)
					bUnzip := false
					if file.Size > 0 {
						header.Set("Content-Length", strconv.Itoa(file.Size))
						if header.Get("Content-Encoding") == "" && strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") {
							header.Set("Content-Encoding", "gzip")
						} else {
							bUnzip = true
						}
					} else {
						header.Set("Content-Length", strconv.Itoa(len(file.Data)))
					}
					c.Response().WriteHeader(http.StatusOK)
					if bUnzip {
						reader, err := gzip.NewReader(strings.NewReader(file.Data))
						if err != nil {
							return err
						}
						defer reader.Close()
						io.Copy(rw, reader)
					} else {
						io.WriteString(rw, file.Data)
					}
					return nil
				}
			}
		}
	}
	e.GET(prefix, h)
	if prefix == "/" {
		e.GET(prefix+"*", h)
	} else {
		e.GET(prefix+"/*", h)
	}
}
