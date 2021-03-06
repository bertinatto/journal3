package http

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	journal "github.com/bertinatto/journal3"
	"github.com/gorilla/mux"
)

func (s *Server) handlePostEdit(w http.ResponseWriter, r *http.Request) {
	permalink, ok := mux.Vars(r)["permalink"]
	if !ok {
		Error(w, r, &journal.Error{Code: journal.EBADINPUT, Message: "Missing permalink"})
		return
	}

	post, err := s.JournalService.FindPostByPermalink(r.Context(), permalink)
	if err != nil {
		Error(w, r, err)
		return
	}

	err = tmpl.ExecuteTemplate(w, "editpost", post)
	if err != nil {
		Error(w, r, err)
		return
	}
}

func (s *Server) handlePostUpdate(w http.ResponseWriter, r *http.Request) {
	permalink, ok := mux.Vars(r)["permalink"]
	if !ok {
		Error(w, r, &journal.Error{Code: journal.EBADINPUT, Message: "Missing permalink"})
		return
	}

	err := r.ParseForm()
	if err != nil {
		Error(w, r, &journal.Error{Code: journal.EBADINPUT, Message: "Failed parsing input"})
	}

	title := strings.TrimSpace(r.Form.Get("title"))
	content := strings.TrimSpace(strings.ReplaceAll(r.Form.Get("content"), "\r\n", "\n"))
	if title == "" || content == "" {
		Error(w, r, &journal.Error{Code: journal.EBADINPUT, Message: "Missing parameters: title and/or content"})
		return
	}

	updatedPost := &journal.PostUpdate{
		Title:   &title,
		Content: &content,
	}

	err = s.JournalService.UpdatePost(r.Context(), permalink, updatedPost)
	if err != nil {
		Error(w, r, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/post/%s", permalink), http.StatusFound)
}

func (s *Server) handlePostCreate(w http.ResponseWriter, r *http.Request) {
	permalink, ok := mux.Vars(r)["permalink"]
	if !ok {
		Error(w, r, &journal.Error{Code: journal.EBADINPUT, Message: "Missing permalink"})
		return
	}

	err := r.ParseForm()
	if err != nil {
		Error(w, r, &journal.Error{Code: journal.EBADINPUT, Message: "Failed parsing input"})
		return
	}

	title := strings.TrimSpace(r.Form.Get("title"))
	content := strings.TrimSpace(strings.ReplaceAll(r.Form.Get("content"), "\r\n", "\n"))
	if title == "" || content == "" {
		Error(w, r, &journal.Error{Code: journal.EBADINPUT, Message: "Missing parameters: title and/or content"})
		return
	}

	post := &journal.Post{
		Permalink: permalink,
		Title:     title,
		Content:   content,
	}

	err = s.JournalService.CreatePost(r.Context(), post)
	if err != nil {
		Error(w, r, err)
	}

	http.Redirect(w, r, fmt.Sprintf("/post/%s", permalink), http.StatusFound)
}

func (s *Server) handlePostView(w http.ResponseWriter, r *http.Request) {
	permalink, ok := mux.Vars(r)["permalink"]
	if !ok {
		Error(w, r, &journal.Error{Code: journal.EBADINPUT, Message: "Missing permalink"})
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
