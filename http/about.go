package http

import "net/http"

func (s *Server) handleAboutCreate(w http.ResponseWriter, r *http.Request) {
}

func (s *Server) handleAboutView(w http.ResponseWriter, r *http.Request) {
	err := tmpl.ExecuteTemplate(w, "about", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
