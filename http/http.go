package http

import (
	"net/http"

	journal "github.com/bertinatto/journal3"
)

var errorCodes = map[string]int{
	journal.ENOTFOUND: http.StatusNotFound,
	journal.EINTERNAL: http.StatusInternalServerError,
}

func ErrorStatusCode(code string) int {
	if v, ok := errorCodes[code]; ok {
		return v
	}
	return http.StatusInternalServerError
}
