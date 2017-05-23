package main

import ()

const (
	SECRET_KEYWORD = "STATIN_CDSS"
	UID_KEY        = "UID_KEY"
	USERID_KEY     = "USERID_KEY"
)

var gConfig struct {
	Port struct {
		Web    string `json:"web"`
		Mobile string `json:"mobile"`
	} `json:"port"`
	StoreUrl string `json:"store_url"`
	StudyId  string `json:"study_id"`
	SSL      struct {
		Enable       bool   `json:"enable"`
		RedirectPort string `json:"redirect_port"`
		Cert         struct {
			Public  string `json:"public"`
			Private string `json:"private"`
		} `json:"cert"`
	} `json:"ssl"`
	Mail struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		UserId   string `json:"userid"`
		Password string `json:"password"`
	} `json:"mail"`
}
