package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
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
	var assetPath string
	var reverseProxies int
	flag.IntVar(&port, "port", 80, "port number")
	flag.IntVar(&reverseProxies, "proxies", 0, "number of reverse proxies")
	flag.StringVar(&configPath, "config", "config.json", "configuration file")
	flag.StringVar(&assetPath, "assets", "assets", "assets directory")

	flag.Parse()

	cfg, err := LoadConfig(configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "load config:", err)
		os.Exit(1)
	}
	server := &Server{
		Config:   cfg,
		AssetDir: assetPath,
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
	http.Handle("/assets/", http.StripPrefix("/assets/",
		http.FileServer(http.Dir(assetPath))))

	if err := http.ListenAndServe(":"+strconv.Itoa(port), nil); err != nil {
		fmt.Fprintln(os.Stderr, "listen:", err)
		os.Exit(1)
	}
}

type Server struct {
	Config     *Config
	AssetDir   string
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
	http.ServeFile(w, r, filepath.Join(s.AssetDir, "index.html"))
}

// Login handles the login system.
func (s *Server) Login(w http.ResponseWriter, r *http.Request) {
	disableCache(w)
	if r.Method == "GET" {
		http.ServeFile(w, r, filepath.Join(s.AssetDir, "login.html"))
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

// GetPrefsAPI serves the API to view preferences.
func (s *Server) GetPrefsAPI(w http.ResponseWriter, r *http.Request) {
	disableCache(w)
	if !s.authOrError(w, r) {
		return
	}
	obj := map[string]interface{}{
		"logSize": s.Config.LogSize(),
	}
	s.servePayload(w, obj)
}

// SetPrefsAPI serves the API to set preferences.
func (s *Server) SetPrefsAPI(w http.ResponseWriter, r *http.Request) {
	disableCache(w)
	if !s.authOrError(w, r) {
		return
	}

	contents, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var prefObj struct {
		LogSize int `json:"logSize"`
	}
	if err := json.Unmarshal(contents, &prefObj); err != nil {
		s.serveError(w, "JSON unmarshal: "+err.Error())
		return
	}
	if err := s.Config.SetLogSize(prefObj.LogSize); err != nil {
		s.serveError(w, "could not save settings")
	} else {
		s.servePayload(w, true)
	}
}

func (s *Server) authOrError(w http.ResponseWriter, r *http.Request) bool {
	if !s.authenticated(r) {
		s.serveError(w, "not authenticated")
		return false
	}
	return true
}

func (s *Server) serveError(w http.ResponseWriter, msg string) {
	pkt := map[string]string{"error": msg}
	data, _ := json.Marshal(pkt)
	w.Write(data)
}

func (s *Server) servePayload(w http.ResponseWriter, msg interface{}) {
	pkt := map[string]interface{}{"data": msg}
	data, err := json.Marshal(pkt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		w.Write(data)
	}
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
