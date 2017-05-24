package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/url"
	"os"

	"git.tnicdev.co.kr/research/statin_cdss/pkg/subject"
	"git.tnicdev.co.kr/research/statin_cdss/pkg/user"

	"github.com/ghodss/yaml"
)

var gAPI = struct {
	SubjectStore *subject.Store
	UserStore    *user.Store
}{}

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

	if false {
		gAPI.UserStore, err = user.NewStore(u, "user")
		if err != nil {
			log.Fatalln(err.Error())
		}
		gAPI.SubjectStore, err = subject.NewStore(u, "subject")
		if err != nil {
			log.Fatalln(err.Error())
		}
	}
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
