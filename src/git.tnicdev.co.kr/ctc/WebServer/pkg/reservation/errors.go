package reservation

import (
	"errors"
)

var (
	ErrNotExistReservation        = errors.New("not exist reservation")
	ErrNotExistReservationSubject = errors.New("not exist reservation subject")
	ErrExistReservationSubject    = errors.New("exist reservation subject")
)
