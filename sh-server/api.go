package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/websocket"
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
		s.Log.LogSizeUpdated()
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

// AddAPI serves the API for adding a log entry.
func (s *Server) AddAPI(w http.ResponseWriter, r *http.Request) {
	var obj struct {
		Service string `json:"service"`
		Message string `json:"message"`
	}
	if !s.processAPICall(w, r, &obj) {
		return
	}
	id, err := s.Log.Add(obj.Service, obj.Message)
	if err != nil {
		s.serveError(w, err.Error())
	} else {
		s.servePayload(w, id)
	}
}

// OverviewAPI serves the API for seeing the log overview.
func (s *Server) OverviewAPI(w http.ResponseWriter, r *http.Request) {
	if !s.processAPICall(w, r, nil) {
		return
	}
	s.serveLog(w, s.Log.Overview())
}

// FullLogAPI serves the API for seeing the entire log.
func (s *Server) FullLogAPI(w http.ResponseWriter, r *http.Request) {
	if !s.processAPICall(w, r, nil) {
		return
	}
	s.serveLog(w, s.Log.FullLog())
}

// ServiceLogAPI serves the API for seeing the log of a
// specific service.
func (s *Server) ServiceLogAPI(w http.ResponseWriter, r *http.Request) {
	var obj struct {
		Service string `json:"service"`
	}
	if !s.processAPICall(w, r, &obj) {
		return
	}
	records, err := s.Log.ServiceLog(obj.Service)
	if err != nil {
		s.serveError(w, err.Error())
	} else {
		s.serveLog(w, records)
	}
}

// DeleteAPI serves the API for deleting services.
func (s *Server) DeleteAPI(w http.ResponseWriter, r *http.Request) {
	var obj struct {
		Service string `json:"service"`
	}
	if !s.processAPICall(w, r, &obj) {
		return
	}
	if err := s.Log.DeleteService(obj.Service); err != nil {
		s.serveError(w, err.Error())
	} else {
		s.servePayload(w, true)
	}
}

// StreamServiceAPI serves a stream of messages.
func (s *Server) StreamAPI(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Service string `json:"service"`
	}
	if !s.processAPICall(w, r, &data) {
		return
	}
	u := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	conn, err := u.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	for {
		ch := s.Log.WaitService(data.Service)
		select {
		case <-ch:
		case <-r.Context().Done():
			return
		}
		entries, err := s.Log.ServiceLog(data.Service)
		if err != nil {
			conn.WriteJSON(map[string]string{"error": err.Error()})
			return
		}
		// TODO: write the latest entries.
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

func (s *Server) serveLog(w http.ResponseWriter, l []LogRecord) {
	if l == nil {
		s.servePayload(w, []LogRecord{})
	} else {
		s.servePayload(w, l)
	}
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
