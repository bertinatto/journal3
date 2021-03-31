package http

import (
	"net/http"

	journal "github.com/bertinatto/journal3"
	"k8s.io/klog/v2"
)

var errorCodes = map[string]int{
	journal.ENOTFOUND: http.StatusNotFound,
	journal.EBADINPUT: http.StatusBadRequest,
	journal.EINTERNAL: http.StatusInternalServerError,
}

func ErrorStatusCode(code string) int {
	if v, ok := errorCodes[code]; ok {
		return v
	}
	return http.StatusInternalServerError
}

func Error(w http.ResponseWriter, r *http.Request, err error) {
	klog.Error(err)
	code, message := journal.ErrorCode(err), journal.ErrorMessage(err)
	w.WriteHeader(ErrorStatusCode(code))
	err = tmpl.ExecuteTemplate(w, "error", &journal.Error{Message: message})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
