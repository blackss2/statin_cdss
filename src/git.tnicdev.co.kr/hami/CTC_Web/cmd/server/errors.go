package main

import (
	"errors"
)

var (
	ErrNotMatchedStudyId = errors.New("not matched studyid")
	ErrNotAuthorized     = errors.New("not authorized")
	ErrNotExistForm      = errors.New("not exist form")
	ErrExistSubject      = errors.New("exist subject")
	ErrExistUser         = errors.New("exist user")
)
