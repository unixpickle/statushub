package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/unixpickle/essentials"
)

const CookieExpration = time.Hour * 24 * 60

type SessionManager struct {
	secret string
}

func NewSessionManager(secret string) *SessionManager {
	if secret == "" {
		data := make([]byte, 16)
		_, err := rand.Read(data)
		essentials.Must(err)
		secret = hex.EncodeToString(data)
	}
	return &SessionManager{
		secret: secret,
	}
}

func (s *SessionManager) CreateSession(w http.ResponseWriter) {
	expire := time.Now().Add(CookieExpration)
	sessionData := strconv.FormatInt(expire.Unix(), 10)
	sessionData = sessionData + "-" + s.signature(sessionData)
	cookie := http.Cookie{
		Name:    "shsess",
		Value:   sessionData,
		Expires: expire,
	}
	http.SetCookie(w, &cookie)
}

func (s *SessionManager) CheckSession(r *http.Request) bool {
	for _, c := range r.Cookies() {
		if c.Name != "shsess" {
			continue
		}
		parts := strings.Split(c.Value, "-")
		if len(parts) != 2 {
			continue
		}
		if parts[1] != s.signature(parts[0]) {
			continue
		}
		date, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			continue
		}
		t := time.Unix(date, 0)
		if time.Now().After(t) {
			continue
		}
		return true
	}
	return false
}

func (s *SessionManager) ClearSession(w http.ResponseWriter) {
	expire := time.Now().Add(CookieExpration)
	cookie := http.Cookie{
		Name:    "shsess",
		Value:   "none",
		Expires: expire,
	}
	http.SetCookie(w, &cookie)
}

func (s *SessionManager) signature(data string) string {
	return hashData(s.secret + data + s.secret)
}

func hashData(data string) string {
	res := sha256.Sum256([]byte(data))
	return hex.EncodeToString(res[:])
}
