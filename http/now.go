package http

import (
	"net/http"

	journal "github.com/bertinatto/journal3"
)

func (s *Server) handleNowCreate(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	location := r.Form.Get("location")
	content := r.Form.Get("content")
	if location == "" || content == "" {
		http.Error(w, "Bad request: location and content must be present", http.StatusBadRequest)
		return
	}

	now := &journal.Now{
		Content:      content,
		FromLocation: location,
	}

	err = s.NowService.CreateNow(r.Context(), now)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleNowView(w http.ResponseWriter, r *http.Request) {
	now, err := s.NowService.FindLatestNow(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, "now", now)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
