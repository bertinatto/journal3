package http

import (
	"errors"
	"net/http"

	journal "github.com/bertinatto/journal3"
)

func (s *Server) handleNowCreate(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		Error(w, r, &journal.Error{Code: journal.EBADINPUT, Message: "Failed parsing input"})
		return
	}

	location := r.Form.Get("location")
	content := r.Form.Get("content")
	if location == "" || content == "" {
		Error(w, r, &journal.Error{Code: journal.EBADINPUT, Message: "Missing parameters: location and/or content"})
		return
	}

	now := &journal.Now{
		Content:      content,
		FromLocation: location,
	}

	err = s.NowService.CreateNow(r.Context(), now)
	if err != nil {
		Error(w, r, err)
		return
	}

	http.Redirect(w, r, "/now", http.StatusFound)
}

func (s *Server) handleNowView(w http.ResponseWriter, r *http.Request) {
	now, err := s.NowService.FindLatestNow(r.Context())
	if err != nil {
		Error(w, r, err)
		return
	}

	err = tmpl.ExecuteTemplate(w, "now", now)
	if err != nil {
		Error(w, r, err)
		return
	}
}

func (s *Server) handleNowEdit(w http.ResponseWriter, r *http.Request) {
	var e *journal.Error
	now, err := s.NowService.FindLatestNow(r.Context())
	if errors.As(err, &e) {
		if e.Code == journal.ENOTFOUND {
			now = &journal.Now{
				FromLocation: "",
				Content:      "",
			}
		}
	} else if err != nil {
		Error(w, r, err)
		return
	}

	err = tmpl.ExecuteTemplate(w, "editnow", now)
	if err != nil {
		Error(w, r, err)
		return
	}
}
