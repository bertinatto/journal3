package http

import (
	"errors"
	"net/http"
	"strings"

	journal "github.com/bertinatto/journal3"
)

func (s *Server) handleContactCreate(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		Error(w, r, &journal.Error{Code: journal.EBADINPUT, Message: "Failed parsing input"})
		return
	}

	content := strings.TrimSpace(strings.ReplaceAll(r.Form.Get("content"), "\r\n", "\n"))
	if content == "" {
		Error(w, r, &journal.Error{Code: journal.EBADINPUT, Message: "Missing parameters: content"})
		return
	}

	page := &journal.Page{
		Name:    "contact",
		Content: content,
	}

	err = s.PageService.CreatePage(r.Context(), page)
	if err != nil {
		Error(w, r, err)
	}

	http.Redirect(w, r, "/contact", http.StatusFound)
}

func (s *Server) handleContactEdit(w http.ResponseWriter, r *http.Request) {
	page, err := s.PageService.FindPageByName(r.Context(), "contact")
	if err != nil {
		Error(w, r, err)
		return
	}

	err = tmpl.ExecuteTemplate(w, "editpage", page)
	if err != nil {
		Error(w, r, err)
		return
	}
}

func (s *Server) handleContactUpdate(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		Error(w, r, &journal.Error{Code: journal.EBADINPUT, Message: "Failed parsing input"})
	}

	content := strings.TrimSpace(strings.ReplaceAll(r.Form.Get("content"), "\r\n", "\n"))
	if content == "" {
		Error(w, r, &journal.Error{Code: journal.EBADINPUT, Message: "Missing parameters: content"})
		return
	}

	updatedPage := &journal.PageUpdate{
		Content: &content,
	}

	err = s.PageService.UpdatePage(r.Context(), "contact", updatedPage)
	if err != nil {
		Error(w, r, err)
		return
	}

	http.Redirect(w, r, "/contact", http.StatusFound)

}

func (s *Server) handleContactView(w http.ResponseWriter, r *http.Request) {
	var e *journal.Error
	page, err := s.PageService.FindPageByName(r.Context(), "contact")
	if errors.As(err, &e) {
		if e.Code == journal.ENOTFOUND {
			err = tmpl.ExecuteTemplate(w, "newpage", "contact")
			if err != nil {
				Error(w, r, err)
				return
			}
			return
		}
	}
	if err != nil {
		Error(w, r, err)
		return
	}

	err = tmpl.ExecuteTemplate(w, "page", page)
	if err != nil {
		Error(w, r, err)
		return
	}
}
