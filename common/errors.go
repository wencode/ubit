package common

import (
	"errors"
)

var (
	ErrInvalidState    = errors.New("ubit: invalid state")
	ErrInvalidArgument = errors.New("ubit: invalid arguments")
)
