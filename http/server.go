package http

import (
	"context"
	"html/template"
	"net"
	"net/http"
	"time"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/parser"
	"github.com/gorilla/mux"
	"k8s.io/klog/v2"

	journal "github.com/bertinatto/journal3"
	"github.com/bertinatto/journal3/http/assets"
	"github.com/bertinatto/journal3/http/html"
)

var tmpl = template.Must(template.New("").Funcs(
	template.FuncMap{
		"safeHTML": func(content string) template.HTML {
			parser := parser.NewWithExtensions(parser.CommonExtensions |
				parser.FencedCode |
				parser.HardLineBreak |
				parser.NoEmptyLineBeforeBlock |
				parser.EmptyLinesBreakList)
			return template.HTML(markdown.ToHTML([]byte(content), parser, nil))
		},
	},
).ParseFS(html.FS, "*.tmpl"))

type Server struct {
	ln     net.Listener
	server *http.Server
	router *mux.Router

	JournalService journal.JournalService
	NowService     journal.NowService
	UserService    journal.UserService
}

func NewServer() *Server {
	s := &Server{
		server: &http.Server{},
		router: mux.NewRouter().StrictSlash(true),
	}
	s.router.Use(s.handlePanic)
	s.router.Use(s.handleMethodOverride)
	s.router.HandleFunc("/", s.handleIndex).Methods("GET")
	s.router.HandleFunc("/about", s.handleAboutView).Methods("GET")
	s.router.HandleFunc("/about", s.handleAboutCreate).Methods("GET")
	s.router.HandleFunc("/now", s.handleNowView).Methods("GET")
	s.router.HandleFunc("/now", s.handleNowCreate).Methods("POST")
	s.router.HandleFunc("/post", s.handlePostView).Methods("GET")
	s.router.HandleFunc("/post/{permalink}", s.handlePostView).Methods("GET")
	s.router.HandleFunc("/post/{permalink}/edit", s.handlePostEdit).Methods("GET")
	s.router.HandleFunc("/post/{permalink}/edit", s.handlePostUpdate).Methods("POST")
	s.router.HandleFunc("/post/{permalink}", s.handlePostCreate).Methods("POST")
	s.router.PathPrefix("/assets").Handler(http.StripPrefix("/assets", http.FileServer(http.FS(assets.FS))))
	s.router.PathPrefix("/uploads").Handler(http.StripPrefix("/uploads", http.FileServer(http.Dir("http/upload/"))))
	s.server.Handler = http.HandlerFunc(s.router.ServeHTTP)
	return s
}

func (s *Server) Open(address string) error {
	// TODO:
	ln, err := net.Listen("tcp", address)
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

func (s *Server) handleMethodOverride(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			method := r.PostFormValue("_method")
			if method == http.MethodPut || method == http.MethodPatch {
				r.Method = method
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	posts, err := s.JournalService.FindPosts(r.Context())
	if err != nil {
		Error(w, r, err)
		return
	}
	err = tmpl.ExecuteTemplate(w, "index", posts)
	if err != nil {
		Error(w, r, err)
		return
	}
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
