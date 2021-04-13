package http

import (
	"fmt"
	"net/http"

	journal "github.com/bertinatto/journal3"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
	"k8s.io/klog/v2"
)

var (
	key           = []byte("super-secret-key")
	store         = sessions.NewCookieStore(key)
	sessionCookie = "journal3-session"
)

func (s *Server) handleSingUpView(w http.ResponseWriter, r *http.Request) {
	err := tmpl.ExecuteTemplate(w, "signup", nil)
	if err != nil {
		Error(w, r, err)
		return
	}
}

func (s *Server) handleSingUp(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		Error(w, r, &journal.Error{Code: journal.EBADINPUT, Message: "Failed parsing input"})
		return
	}

	name := r.Form.Get("name")
	email := r.Form.Get("email")
	password := r.Form.Get("password")
	if name == "" || email == "" || password == "" {
		Error(w, r, &journal.Error{Code: journal.EBADINPUT, Message: "Missing parameters: name, email and passport are required"})
		return
	}

	_, err = s.UserService.FindUserByEmail(r.Context(), email)
	if err == nil {
		Error(w, r, err)
		return
	}

	u := &journal.User{
		Name:     name,
		Email:    email,
		Password: password,
	}

	err = u.Validate()
	if err != nil {
		Error(w, r, &journal.Error{Code: journal.EBADINPUT, Message: fmt.Sprintf("Invalid user: %v", err)})
		return
	}

	// Salt and hash the password using the bcrypt algorithm
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 8)
	if err != nil {
		Error(w, r, &journal.Error{Code: journal.EINTERNAL, Message: "Failed to process user"})
		return
	}
	u.Password = string(hashedPassword)

	err = s.UserService.CreateUser(r.Context(), u)
	if err != nil {
		Error(w, r, err)
		return
	}

	session, err := store.Get(r, sessionCookie)
	if err != nil {
		Error(w, r, err)
		return
	}
	session.Options.MaxAge = 60 * 60 * 24 // 24 hours
	session.Values["authenticated"] = true
	session.Values["uid"] = u.ID
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) handleLoginView(w http.ResponseWriter, r *http.Request) {
	err := tmpl.ExecuteTemplate(w, "login", nil)
	if err != nil {
		Error(w, r, err)
		return
	}
}

func (s *Server) handleLoginCreate(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		Error(w, r, err)
		return
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")
	if email == "" || password == "" {
		Error(w, r, &journal.Error{Code: journal.EBADINPUT, Message: "Missing parameters: email and passport are required"})
		return
	}

	user, err := s.UserService.FindUserByEmail(r.Context(), email)
	if err != nil {
		Error(w, r, err)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		Error(w, r, &journal.Error{Code: journal.ENOTAUTHORIZED, Message: "Invalid password"})
		return
	}

	session, err := store.Get(r, sessionCookie)
	if err != nil {
		Error(w, r, err)
		return
	}

	session.Options.MaxAge = 60 * 60 * 24 // 24 hours
	session.Values["authenticated"] = true
	session.Values["uid"] = user.ID
	session.Save(r, w)

	redirect, ok := session.Values["redirect"].(string)
	if !ok {
		redirect = "/"
	}

	http.Redirect(w, r, redirect, http.StatusSeeOther)
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, sessionCookie)
	if err != nil {
		Error(w, r, err)
		return
	}
	session.Options.MaxAge = -1
	session.Values["authenticated"] = false
	session.Values["uid"] = 0
	session.Save(r, w)
	klog.Infof("Session logout %+v", session)
	http.Redirect(w, r, "/", http.StatusFound)

}
