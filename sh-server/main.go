package main

import (
	"flag"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/context"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/unixpickle/essentials"
	"github.com/unixpickle/ratelimit"
)

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
		essentials.Die("load config:", err)
	}
	server := &Server{
		Config: cfg,
		Log:    NewLog(cfg),
		Sessions: sessions.NewCookieStore(securecookie.GenerateRandomKey(16),
			securecookie.GenerateRandomKey(16)),
		LoginLimit: ratelimit.NewTimeSliceLimiter(RateLimitDuration, RateLimitAttempts),
		LimitNamer: &ratelimit.HTTPRemoteNamer{NumProxies: reverseProxies},
	}

	handlers := map[string]http.HandlerFunc{
		"/":                  server.Root,
		"/login":             server.Login,
		"/logout":            server.Logout,
		"/api/getprefs":      server.GetPrefsAPI,
		"/api/setprefs":      server.SetPrefsAPI,
		"/api/chpass":        server.ChpassAPI,
		"/api/add":           server.AddAPI,
		"/api/addBatch":      server.AddBatchAPI,
		"/api/addMedia":      server.AddMediaAPI,
		"/api/overview":      server.OverviewAPI,
		"/api/mediaOverview": server.MediaOverviewAPI,
		"/api/fullLog":       server.FullLogAPI,
		"/api/serviceLog":    server.ServiceLogAPI,
		"/api/mediaLog":      server.MediaLogAPI,
		"/api/mediaView":     server.MediaViewAPI,
		"/api/delete":        server.DeleteAPI,
		"/api/deleteMedia":   server.DeleteMediaAPI,
		"/api/serviceStream": server.ServiceStreamAPI,
		"/api/fullStream":    server.FullStreamAPI,
	}
	for path, f := range handlers {
		http.Handle(path, context.ClearHandler(f))
	}
	http.Handle("/assets/", http.StripPrefix("/assets/",
		http.FileServer(assetFS())))

	if err := http.ListenAndServe(":"+strconv.Itoa(port), nil); err != nil {
		essentials.Die("listen:", err)
	}
}

type Server struct {
	Config     *Config
	Log        *Log
	Sessions   *sessions.CookieStore
	LoginLimit *ratelimit.TimeSliceLimiter
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
	limitID := s.LimitNamer.Name(r)
	if s.LoginLimit.Get(limitID) < 0 {
		http.Error(w, "too many login attempts", http.StatusTooManyRequests)
		return
	}
	pass := r.FormValue("password")
	if !s.Config.CheckPass(pass) {
		s.LoginLimit.Decrement(limitID)
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
