package main

import (
	"errors"
)

var (
	ErrNotMatchedStudyId = errors.New("not matched studyid")
	ErrNotAuthorized     = errors.New("not authorized")
	ErrNotExistForm      = errors.New("not exist form")
	ErrNotExistGroup     = errors.New("not exist group")
	ErrExistSubject      = errors.New("exist subject")
	ErrExistUser         = errors.New("exist user")
)
