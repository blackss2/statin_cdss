package user

import (
	"errors"
)

var (
	ErrExistUserId     = errors.New("exist userid")
	ErrNotExistUser    = errors.New("not exist user")
	ErrInvalidPassword = errors.New("invalid password")
)
