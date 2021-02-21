package http

import (
	"net"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"k8s.io/klog/v2"

	journal "github.com/bertinatto/journal3"
)

type Server struct {
	ln     net.Listener
	server *http.Server
	router *mux.Router

	JournalService journal.JournalService
}

func NewServer() *Server {
	s := &Server{
		server: &http.Server{},
		router: mux.NewRouter(),
	}

	s.server.Handler = http.HandlerFunc(s.serveHTTP)

	s.router.HandleFunc("/", s.handleIndex).Methods("GET")
	s.router.HandleFunc("/now", s.handleNow).Methods("GET")
	s.router.HandleFunc("/about", s.handleAbout).Methods("GET")
	s.router.HandleFunc("/post", s.handlePostCreate).Methods("POST")
	s.router.HandleFunc("/post/{id}", s.handlePostView).Methods("GET")

	return s
}

func (s *Server) Open() error {
	ln, err := net.Listen("tcp", "127.0.0.1:1111")
	if err != nil {
		return err
	}
	s.ln = ln
	go s.server.Serve(s.ln)
	return nil
}

func (s *Server) serveHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("You've reached the index page!"))
}

func (s *Server) handleNow(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("You've reached the /now page!"))
}

func (s *Server) handleAbout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("You've reached the /about page!"))
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

	klog.Infof("Inserting %+v\n", post)
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

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(post.Title))
	w.Write([]byte("\n"))
	w.Write([]byte(post.Content))
}
