package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// GetPrefsAPI serves the API to view preferences.
func (s *Server) GetPrefsAPI(w http.ResponseWriter, r *http.Request) {
	if !s.processAPICall(w, r, nil) {
		return
	}
	obj := map[string]interface{}{
		"logSize": s.Config.LogSize(),
	}
	s.servePayload(w, obj)
}

// SetPrefsAPI serves the API to set preferences.
func (s *Server) SetPrefsAPI(w http.ResponseWriter, r *http.Request) {
	var prefObj struct {
		LogSize int `json:"logSize"`
	}
	if !s.processAPICall(w, r, &prefObj) {
		return
	}
	if err := s.Config.SetLogSize(prefObj.LogSize); err != nil {
		s.serveError(w, "could not save settings")
	} else {
		s.servePayload(w, true)
	}
}

// ChpassAPI serves the API for password changing.
func (s *Server) ChpassAPI(w http.ResponseWriter, r *http.Request) {
	var obj struct {
		Old     string `json:"old"`
		Confirm string `json:"confirm"`
		New     string `json:"new"`
	}
	if !s.processAPICall(w, r, &obj) {
		return
	}
	if obj.Old != obj.Confirm {
		s.serveError(w, "passwords do not match")
		return
	}
	if !s.Config.CheckPass(obj.Old) {
		s.serveError(w, "password incorrect")
		return
	}
	if err := s.Config.SetPass(obj.New); err != nil {
		s.serveError(w, "could not save settings")
	} else {
		s.servePayload(w, true)
	}
}

func (s *Server) processAPICall(w http.ResponseWriter, r *http.Request, inData interface{}) bool {
	disableCache(w)
	if !s.authenticated(r) {
		s.serveError(w, "not authenticated")
		return false
	}
	if inData != nil {
		contents, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return false
		}
		if err := json.Unmarshal(contents, inData); err != nil {
			s.serveError(w, "JSON unmarshal: "+err.Error())
			return false
		}
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
