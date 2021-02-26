package http

import (
	"context"
	"html/template"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"k8s.io/klog/v2"

	journal "github.com/bertinatto/journal3"
	"github.com/bertinatto/journal3/http/assets"
	"github.com/bertinatto/journal3/http/html"
)

var tmpl = template.Must(template.ParseFS(html.FS, "*.tmpl"))

type Server struct {
	ln     net.Listener
	server *http.Server
	router *mux.Router

	JournalService journal.JournalService
	NowService     journal.NowService
}

func NewServer() *Server {
	s := &Server{
		server: &http.Server{},
		router: mux.NewRouter(),
	}

	s.router.Use(s.handlePanic)

	s.router.HandleFunc("/", s.handleIndex).Methods("GET")
	// s.router.HandleFunc("/contact", s.handleContact).Methods("GET")
	s.router.HandleFunc("/now", s.handleNowView).Methods("GET")
	s.router.HandleFunc("/now", s.handleNowCreate).Methods("POST")

	s.router.HandleFunc("/about", s.handleAbout).Methods("GET")
	s.router.HandleFunc("/post/{id}", s.handlePostView).Methods("GET")
	s.router.HandleFunc("/post", s.handlePostCreate).Methods("POST")

	s.router.PathPrefix("/assets").Handler(http.StripPrefix("/assets", http.FileServer(http.FS(assets.FS))))

	s.server.Handler = http.HandlerFunc(s.router.ServeHTTP)
	return s
}

func (s *Server) Open() error {
	// TODO: why am I using a listener instead of calling ListenAndServe()?
	ln, err := net.Listen("tcp", "127.0.0.1:1111")
	if err != nil {
		return err
	}
	s.ln = ln
	go s.server.Serve(s.ln)
	return nil
}

func (s *Server) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	return s.server.Shutdown(ctx)
}

func (s *Server) serveHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) handlePanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				e := err.(error)
				// TODO: report panic to sentry/rollbar
				w.Write([]byte(e.Error()))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleAbout(w http.ResponseWriter, r *http.Request) {
	err := tmpl.ExecuteTemplate(w, "now", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) handlePostCreate(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	title := r.Form.Get("title")
	content := r.Form.Get("content")
	if title == "" || content == "" {
		http.Error(w, "Bad request: title and content must be present", http.StatusBadRequest)
		return
	}

	post := &journal.Post{
		Title:   title,
		Content: content,
	}

	err = s.JournalService.CreatePost(r.Context(), post)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// TODO: how can I get debug logs more effectively
	klog.V(5).Infof("Inserting %+v\n", post)
}

func (s *Server) handlePostView(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	post, err := s.JournalService.FindPostByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, "post", post)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	posts, err := s.JournalService.FindPosts(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, "index", posts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

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

	// TODO: how can I get debug logs more effectively
	klog.V(5).Infof("Inserting %+v\n", now)
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
