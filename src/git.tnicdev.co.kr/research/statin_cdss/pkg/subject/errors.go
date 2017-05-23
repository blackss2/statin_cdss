package subject

import (
	"errors"
)

var (
	ErrExistSubjectId  = errors.New("exist subject_id")
	ErrNotExistSubject = errors.New("not exist subject")
)
