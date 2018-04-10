package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/unixpickle/statushub"
)

// GetPrefsAPI serves the API to view preferences.
func (s *Server) GetPrefsAPI(w http.ResponseWriter, r *http.Request) {
	if !s.processAPICall(w, r, nil) {
		return
	}
	obj := map[string]interface{}{
		"logSize":    s.Config.LogSize(),
		"mediaCache": s.Config.MediaCache(),
	}
	s.servePayload(w, obj)
}

// SetPrefsAPI serves the API to set preferences.
func (s *Server) SetPrefsAPI(w http.ResponseWriter, r *http.Request) {
	var prefObj struct {
		LogSize    int `json:"logSize"`
		MediaCache int `json:"mediaCache"`
	}
	if !s.processAPICall(w, r, &prefObj) {
		return
	}

	if err := s.Config.SetLogSize(prefObj.LogSize); err != nil {
		s.serveError(w, "could not set log size")
		return
	}
	s.Log.LogSizeUpdated()

	if err := s.Config.SetMediaCache(prefObj.MediaCache); err != nil {
		s.serveError(w, "could not set media cache")
		return
	}
	s.Log.MediaCacheUpdated()

	s.servePayload(w, true)
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
	if obj.New != obj.Confirm {
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

// AddMediaAPI serves the API for adding a media entry.
func (s *Server) AddMediaAPI(w http.ResponseWriter, r *http.Request) {
	var obj struct {
		Folder   string `json:"folder"`
		Filename string `json:"filename"`
		Mime     string `json:"mime"`
		Data     []byte `json:"data"`
	}
	if !s.processAPICall(w, r, &obj) {
		return
	}
	id, err := s.Log.AddMedia(obj.Folder, obj.Filename, obj.Mime, obj.Data)
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

// MediaOverviewAPI serves the API for seeing the media
// overview.
func (s *Server) MediaOverviewAPI(w http.ResponseWriter, r *http.Request) {
	if !s.processAPICall(w, r, nil) {
		return
	}
	s.serveMediaLog(w, s.Log.MediaOverview())
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

// MediaLogAPI serves the API for seeing the log of a
// media folder.
func (s *Server) MediaLogAPI(w http.ResponseWriter, r *http.Request) {
	var obj struct {
		Folder string `json:"folder"`
	}
	if !s.processAPICall(w, r, &obj) {
		return
	}
	records, err := s.Log.MediaLog(obj.Folder)
	if err != nil {
		s.serveError(w, err.Error())
	} else {
		s.serveMediaLog(w, records)
	}
}

// MediaAPI serves the contents of a media item.
func (s *Server) MediaViewAPI(w http.ResponseWriter, r *http.Request) {
	disableCache(w)
	if !s.authenticated(r) {
		http.Error(w, "not authenticated", http.StatusForbidden)
		return
	}
	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	record := s.Log.MediaRecord(id)
	if record == nil {
		http.Error(w, "unknown media record", http.StatusNotFound)
	} else {
		disposition := "inline; filename*=UTF-8''" + url.PathEscape(record.Filename)
		w.Header().Set("Content-Disposition", disposition)
		w.Header().Set("Content-Type", record.Mime)
		http.ServeContent(w, r, record.Filename, time.Unix(record.Time, 0),
			bytes.NewReader(record.Data))
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

// DeleteMediaAPI serves the API for deleting media.
func (s *Server) DeleteMediaAPI(w http.ResponseWriter, r *http.Request) {
	var obj struct {
		Folder string `json:"folder"`
	}
	if !s.processAPICall(w, r, &obj) {
		return
	}
	if err := s.Log.DeleteMedia(obj.Folder); err != nil {
		s.serveError(w, err.Error())
	} else {
		s.servePayload(w, true)
	}
}

// ServiceStreamAPI serves a stream of messages for a
// particular service.
func (s *Server) ServiceStreamAPI(w http.ResponseWriter, r *http.Request) {
	if !s.authenticated(r) {
		s.serveError(w, "not authenticated")
		return
	}
	service := r.FormValue("service")
	s.serveStream(w, r, func() <-chan struct{} {
		return s.Log.WaitService(service)
	}, func() []statushub.LogRecord {
		res, _ := s.Log.ServiceLog(service)
		return res
	})
}

// FullStreamAPI serves a stream of messages for all
// services.
func (s *Server) FullStreamAPI(w http.ResponseWriter, r *http.Request) {
	if !s.authenticated(r) {
		s.serveError(w, "not authenticated")
		return
	}
	s.serveStream(w, r, func() <-chan struct{} {
		return s.Log.Wait()
	}, func() []statushub.LogRecord {
		return s.Log.FullLog()
	})
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

func (s *Server) serveLog(w http.ResponseWriter, l []statushub.LogRecord) {
	if l == nil {
		s.servePayload(w, []statushub.LogRecord{})
	} else {
		s.servePayload(w, l)
	}
}

func (s *Server) serveMediaLog(w http.ResponseWriter, l []statushub.MediaRecord) {
	if l == nil {
		s.servePayload(w, []statushub.MediaRecord{})
	} else {
		s.servePayload(w, l)
	}
}

func (s *Server) serveStream(w http.ResponseWriter, r *http.Request, getWait func() <-chan struct{},
	getEntries func() []statushub.LogRecord) {
	u := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	conn, err := u.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	connDead := make(chan struct{})
	go func() {
		var obj interface{}
		for {
			if conn.ReadJSON(&obj) != nil {
				close(connDead)
				return
			}
		}
	}()

	greatestID := 0
	first := true
	for {
		ch := getWait()
		entries := getEntries()
		if first {
			first = false
			if len(entries) == 0 {
				greatestID = -1
			} else {
				greatestID = entries[0].ID
			}
		} else if len(entries) == 0 {
			// The log was cleared.
			greatestID = -1
		} else {
			startIdx := -1
			for startIdx+1 < len(entries) && entries[startIdx+1].ID > greatestID {
				startIdx++
			}
			for i := startIdx; i >= 0; i-- {
				if conn.WriteJSON(entries[i]) != nil {
					return
				}
			}
			greatestID = entries[0].ID
		}
		select {
		case <-ch:
		case <-connDead:
			return
		}
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
