package http

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	journal "github.com/bertinatto/journal3"
	"github.com/gorilla/mux"
)

func (s *Server) handlePostCreate(w http.ResponseWriter, r *http.Request) {
	permalink, ok := mux.Vars(r)["permalink"]
	if !ok {
		Error(w, r, fmt.Errorf("failed to parse permalink"))
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	title := strings.TrimSpace(r.Form.Get("title"))
	content := strings.TrimSpace(strings.ReplaceAll(r.Form.Get("content"), "\r\n", "\n"))
	if title == "" || content == "" {
		http.Error(w, "Bad request: title, content and permalink must be present", http.StatusBadRequest)
		return
	}

	post := &journal.Post{
		Permalink: permalink,
		Title:     title,
		Content:   content,
	}

	err = s.JournalService.CreatePost(r.Context(), post)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	http.Redirect(w, r, fmt.Sprintf("/post/%s", permalink), http.StatusMovedPermanently)
}

func (s *Server) handlePostView(w http.ResponseWriter, r *http.Request) {
	permalink, ok := mux.Vars(r)["permalink"]
	if !ok {
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
		return
	}

	var e *journal.Error
	post, err := s.JournalService.FindPostByPermalink(r.Context(), permalink)
	if errors.As(err, &e) {
		if e.Code == journal.ENOTFOUND {
			err = tmpl.ExecuteTemplate(w, "newpost", permalink)
			if err != nil {
				Error(w, r, err)
				return
			}
		}
		return
	}
	if err != nil {
		Error(w, r, err)
		return
	}

	err = tmpl.ExecuteTemplate(w, "post", post)
	if err != nil {
		Error(w, r, err)
		return
	}
}
