package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/url"
	"os"

	"git.tnicdev.co.kr/research/statin_cdss/pkg/notice"
	"git.tnicdev.co.kr/research/statin_cdss/pkg/reservation"
	"git.tnicdev.co.kr/research/statin_cdss/pkg/user"

	"github.com/dancannon/gorethink"
	"github.com/ghodss/yaml"
)

var gAPI = struct {
	SubjectTable     *SubjectTable
	StackTable       *StackTable
	VisitTable       *VisitTable
	FormTable        *FormTable
	DataTable        *DataTable
	HistoryTable     *HistoryTable
	UserStore        *user.Store
	NoticeStore      *notice.Store
	ReservationStore *reservation.Store
}{}

var gFormListCache []*Form

func init() {
	file, err := os.Open("./config.yaml")
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatalln(err)
	}

	err = yaml.Unmarshal(data, &gConfig)
	if err != nil {
		log.Fatalln(err)
	}

	u, err := url.Parse(gConfig.StoreUrl)
	if err != nil {
		log.Fatalln(err)
	}

	session, err := gorethink.Connect(gorethink.ConnectOpts{
		Address: u.Host,
	})
	if err != nil {
		log.Fatalln(err.Error())
	}
	db := createDB(session, u.Path[1:])

	gAPI.SubjectTable = NewSubjectTable(session, db)
	gAPI.StackTable = NewStackTable(session, db)
	gAPI.VisitTable = NewVisitTable(session, db)
	gAPI.FormTable = NewFormTable(session, db)
	gAPI.DataTable = NewDataTable(session, db)
	gAPI.HistoryTable = NewHistoryTable(session, db)

	gAPI.UserStore, err = user.NewStore(u, "user")
	if err != nil {
		log.Fatalln(err.Error())
	}
	gAPI.NoticeStore, err = notice.NewStore(u, "notice")
	if err != nil {
		log.Fatalln(err.Error())
	}
	gAPI.ReservationStore, err = reservation.NewStore(u, "reservation")
	if err != nil {
		log.Fatalln(err.Error())
	}
}

func FormWithCache(ft *FormTable, StudyId string) ([]*Form, error) {
	if gFormListCache == nil {
		forms, err := ft.List(StudyId)
		if err != nil {
			return nil, err
		}
		gFormListCache = forms
	}
	return gFormListCache, nil
}

type Result struct {
	Error  error
	Result interface{}
}

type resultJSON struct {
	Error  interface{} `json:"error"`
	Result interface{} `json:"result"`
}

func (r *Result) MarshalJSON() ([]byte, error) {
	var errorStr interface{}
	if r.Error != nil {
		errorStr = r.Error.Error()
	}
	return json.Marshal(&resultJSON{
		Error:  errorStr,
		Result: r.Result,
	})
}
