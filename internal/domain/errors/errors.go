package errors

import "errors"

var (
	ErrBotNotFound = errors.New("That bot cannot be found in this server.")
)
