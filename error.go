package journal

import (
	"errors"
	"fmt"
)

const (
	ENOTFOUND      = "not_fount"
	EBADINPUT      = "bad_input"
	EINTERNAL      = "internal"
	ENOTAUTHORIZED = "not_authorized"
)

type Error struct {
	Code    string
	Message string
}

func (e *Error) Error() string {
	return fmt.Sprintf("journal error: code %s message %s", e.Code, e.Message)
}

func ErrorCode(err error) string {
	if err == nil {
		return ""
	}

	var e *Error
	if errors.As(err, &e) {
		return e.Code
	}

	return EINTERNAL
}

func ErrorMessage(err error) string {
	if err == nil {
		return ""
	}

	var e *Error
	if errors.As(err, &e) {
		return e.Message
	}

	return "Internal Error"
}
