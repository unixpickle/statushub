package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/unixpickle/ratelimit"
)

type Limiter interface {
	Limit(ipID string) bool
}

const (
	RateLimitDuration = time.Minute * 30
	RateLimitAttempts = 200
)

func main() {
	var port int
	var configPath string
	var reverseProxies int
	flag.IntVar(&port, "port", 80, "port number")
	flag.IntVar(&reverseProxies, "proxies", 0, "number of reverse proxies")
	flag.StringVar(&configPath, "config", "config.json", "configuration file")

	flag.Parse()

	cfg, err := LoadConfig(configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "load config:", err)
		os.Exit(1)
	}
	server := &Server{
		Config: cfg,
		Log:    NewLog(cfg),
		Sessions: sessions.NewCookieStore(securecookie.GenerateRandomKey(16),
			securecookie.GenerateRandomKey(16)),
		LoginLimit: ratelimit.NewTimeSliceLimiter(RateLimitDuration, RateLimitAttempts),
		LimitNamer: &ratelimit.HTTPRemoteNamer{NumProxies: reverseProxies},
	}

	http.HandleFunc("/", server.Root)
	http.HandleFunc("/login", server.Login)
	http.HandleFunc("/logout", server.Logout)
	http.HandleFunc("/api/getprefs", server.GetPrefsAPI)
	http.HandleFunc("/api/setprefs", server.SetPrefsAPI)
	http.HandleFunc("/api/chpass", server.ChpassAPI)
	http.HandleFunc("/api/add", server.AddAPI)
	http.HandleFunc("/api/overview", server.OverviewAPI)
	http.HandleFunc("/api/serviceLog", server.ServiceLogAPI)
	http.HandleFunc("/api/fullLog", server.FullLogAPI)
	http.Handle("/assets/", http.StripPrefix("/assets/",
		http.FileServer(assetFS())))

	if err := http.ListenAndServe(":"+strconv.Itoa(port), nil); err != nil {
		fmt.Fprintln(os.Stderr, "listen:", err)
		os.Exit(1)
	}
}

type Server struct {
	Config     *Config
	Log        *Log
	Sessions   *sessions.CookieStore
	LoginLimit Limiter
	LimitNamer *ratelimit.HTTPRemoteNamer
}

// Root serves the homepage.
func (s *Server) Root(w http.ResponseWriter, r *http.Request) {
	disableCache(w)
	if r.URL.Path != "" && r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if !s.authenticated(r) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	data, _ := Asset("assets/index.html")
	w.Write(data)
}

// Login handles the login system.
func (s *Server) Login(w http.ResponseWriter, r *http.Request) {
	disableCache(w)
	if r.Method == "GET" {
		w.Header().Set("Content-Type", "text/html")
		data, _ := Asset("assets/login.html")
		w.Write(data)
		return
	}
	if s.LoginLimit.Limit(s.LimitNamer.Name(r)) {
		http.Error(w, "too many login attempts", http.StatusTooManyRequests)
		return
	}
	pass := r.FormValue("password")
	if !s.Config.CheckPass(pass) {
		http.Redirect(w, r, "/login?status=failure", http.StatusSeeOther)
		return
	}
	sess, _ := s.Sessions.Get(r, "sessid")
	sess.Values["authenticated"] = true
	sess.Save(r, w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Logout serves the logout function.
func (s *Server) Logout(w http.ResponseWriter, r *http.Request) {
	disableCache(w)
	sess, _ := s.Sessions.Get(r, "sessid")
	sess.Values["authenticated"] = false
	sess.Save(r, w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) authenticated(r *http.Request) bool {
	sess, _ := s.Sessions.Get(r, "sessid")
	val, _ := sess.Values["authenticated"].(bool)
	return val
}

func disableCache(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
}
