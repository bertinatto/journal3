package http

import (
	"context"
	"html/template"
	"net"
	"net/http"
	_ "net/http/pprof"
	"strings"
	"time"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/parser"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/crypto/acme/autocert"
	"k8s.io/klog/v2"

	journal "github.com/bertinatto/journal3"
	"github.com/bertinatto/journal3/http/assets"
	"github.com/bertinatto/journal3/http/html"
)

var (
	requestCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_request_count",
		Help: "Total number of requests by route",
	}, []string{"method", "path"})

	requestSeconds = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_request_seconds",
		Help: "Total amount of request time by route, in seconds",
	}, []string{"method", "path"})
)

var tmpl = template.Must(template.New("").Funcs(
	template.FuncMap{
		"toTitle": func(content string) template.HTML {
			return template.HTML(strings.Title(content))
		},
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

	Domain string
	Addr   string

	PageService    journal.PageService
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
	s.router.NotFoundHandler = http.HandlerFunc(s.handleNotFound)

	s.router.PathPrefix("/assets").Handler(http.StripPrefix("/assets", http.FileServer(http.FS(assets.FS))))
	s.router.PathPrefix("/uploads").Handler(http.StripPrefix("/uploads", http.FileServer(http.Dir("http/upload/"))))

	// Public-facing endopoints, except assets and uploads
	router := s.router.PathPrefix("/").Subrouter()
	router.Use(s.handleMethodOverride)
	router.Use(s.handleSession)
	router.Use(trackMetrics)
	router.HandleFunc("/", s.handleIndex).Methods(http.MethodGet)
	router.HandleFunc("/about", s.handleAboutView).Methods(http.MethodGet)
	router.HandleFunc("/contact", s.handleContactView).Methods(http.MethodGet)
	router.HandleFunc("/now", s.handleNowView).Methods(http.MethodGet)
	router.HandleFunc("/post/{permalink}", s.handlePostView).Methods(http.MethodGet)

	// Register routes that require the user to NOT be authenticated
	{
		r := router.PathPrefix("/").Subrouter()
		r.Use(s.handleNoAuth)
		r.HandleFunc("/signup", s.handleSingUpView).Methods(http.MethodGet)
		r.HandleFunc("/signup", s.handleSingUp).Methods(http.MethodPost)
		r.HandleFunc("/login", s.handleLoginView).Methods(http.MethodGet)
		r.HandleFunc("/login", s.handleLoginCreate).Methods(http.MethodPost)
	}

	// Register routes that require authentication
	{
		r := router.PathPrefix("/").Subrouter()
		r.Use(s.handleAuth)
		r.HandleFunc("/logout", s.handleLogout).Methods(http.MethodGet)
		r.HandleFunc("/about", s.handleAboutCreate).Methods(http.MethodPost)
		r.HandleFunc("/about/edit", s.handleAboutEdit).Methods(http.MethodGet)
		r.HandleFunc("/about", s.handleAboutUpdate).Methods(http.MethodPatch)
		r.HandleFunc("/contact", s.handleContactCreate).Methods(http.MethodPost)
		r.HandleFunc("/contact/edit", s.handleContactEdit).Methods(http.MethodGet)
		r.HandleFunc("/contact", s.handleContactUpdate).Methods(http.MethodPatch)
		r.HandleFunc("/now", s.handleNowCreate).Methods(http.MethodPost)
		r.HandleFunc("/now/edit", s.handleNowEdit).Methods(http.MethodGet)
		r.HandleFunc("/post/{permalink}/edit", s.handlePostEdit).Methods(http.MethodGet)
		r.HandleFunc("/post/{permalink}/edit", s.handlePostUpdate).Methods(http.MethodPatch)
		r.HandleFunc("/post/{permalink}", s.handlePostCreate).Methods(http.MethodPost)
	}

	s.server.Handler = http.HandlerFunc(s.router.ServeHTTP)
	return s
}

func (s *Server) Open() error {
	if s.TLS() {
		s.ln = autocert.NewListener(s.Domain)
	} else {
		l, err := net.Listen("tcp", s.Addr)
		if err != nil {
			return err
		}
		s.ln = l
	}
	go s.server.Serve(s.ln)
	return nil
}

func (s *Server) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	return s.server.Shutdown(ctx)
}

func (s *Server) TLS() bool {
	return s.Domain != ""
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

func (s *Server) handleSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Most requests won't have a session
		session, _ := store.Get(r, sessionCookie)
		if auth, ok := session.Values["authenticated"].(bool); ok && auth {
			if id, ok := session.Values["uid"].(int); ok && id > 0 {
				user, err := s.UserService.FindUserByID(r.Context(), id)
				if err != nil {
					klog.Errorf("Could not find user %d: %v", err)
				} else {
					r = r.WithContext(journal.NewContextWithUser(r.Context(), user))
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleNoAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := journal.UserIDFromContext(r.Context())
		if userID > 0 {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := journal.UserIDFromContext(r.Context())
		if userID > 0 {
			next.ServeHTTP(w, r)
			return
		}

		redirect := "/"
		if r.URL.Path != "" {
			redirect = r.URL.Path
		}

		session, _ := store.Get(r, sessionCookie)
		session.Values["redirect"] = redirect
		session.Save(r, w)

		http.Redirect(w, r, "/login", http.StatusFound)
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

func (s *Server) handleNotFound(w http.ResponseWriter, r *http.Request) {
	err := tmpl.ExecuteTemplate(w, "notfound", nil)
	if err != nil {
		Error(w, r, err)
		return
	}
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

func trackMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := time.Now()

		// Get template path from route
		var tmpl string
		route := mux.CurrentRoute(r)
		if route != nil {
			tmpl, _ = route.GetPathTemplate()
		}

		next.ServeHTTP(w, r)

		// Track total time if template path is valid
		if tmpl != "" {
			requestCount.WithLabelValues(r.Method, tmpl).Inc()
			requestSeconds.WithLabelValues(r.Method, tmpl).Add(float64(time.Since(t).Seconds()))
		}
	})
}

func ListenAndServerTLSRedirect(domain string) error {
	return http.ListenAndServe(":80", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://"+domain, http.StatusFound)
	}))
}

func ListenAndServeDebug() error {
	r := mux.NewRouter().StrictSlash(true)
	r.PathPrefix("/debug/").Handler(http.DefaultServeMux)
	r.Handle("/metrics", promhttp.Handler())
	return http.ListenAndServe(":6060", r)
}
