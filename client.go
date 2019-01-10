// Package statushub is a client for the StatusHub API.
//
// StatusHub is a service for consolidating log messages.
// Essentially, you can have all of your long-running
// scripts write to a single StatusHub host so that you
// can easily check on their outputs.
package statushub

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"

	"github.com/gorilla/websocket"
	"github.com/unixpickle/essentials"
)

// A LogRecord is one logged message.
type LogRecord struct {
	Service string `json:"serviceName"`
	Message string `json:"message"`
	Time    int64  `json:"time"`
	ID      int    `json:"id"`
}

// A MediaRecord is a piece of media stored on the server.
type MediaRecord struct {
	Folder   string `json:"folder"`
	Filename string `json:"filename"`
	Mime     string `json:"mime"`
	Time     int64  `json:"time"`
	ID       int    `json:"id"`
}

// A Client interfaces with a StatusHub back-end.
type Client struct {
	c       *http.Client
	rootURL url.URL
}

// NewClient creates a new, unauthenticated client.
//
// The rootURL specifies the base URL of the StatusHub
// server.
// For example, it might be "http://localhost:8080".
func NewClient(rootURL string) (*Client, error) {
	j, err := cookiejar.New(nil)
	if err != nil {
		return nil, essentials.AddCtx("create cookie jar", err)
	}
	u, err := url.Parse(rootURL)
	if err != nil {
		return nil, essentials.AddCtx("bad root URL", err)
	}
	return &Client{
		c: &http.Client{
			Jar: j,
		},
		rootURL: *u,
	}, nil
}

// Login attempts to authenticate with the server.
func (c *Client) Login(password string) error {
	u := c.rootURL
	u.Path = "/login"
	query := bytes.NewReader([]byte("password=" + url.QueryEscape(password)))
	res, err := c.c.Post(u.String(), "application/x-www-form-urlencoded", query)
	if res != nil {
		res.Body.Close()
	}
	if err != nil {
		return err
	}
	if res.Request.URL.Path == u.Path {
		return errors.New("login failed")
	}
	return nil
}

// Add adds a log record and returns its ID.
func (c *Client) Add(service, message string) (int, error) {
	msg := map[string]string{
		"service": service,
		"message": message,
	}
	var resID int
	err := c.apiCall("add", msg, &resID)
	if err != nil {
		err = essentials.AddCtx("add log record", err)
	}
	return resID, err
}

// AddBatch adds a batch of log records and returns their
// IDs.
func (c *Client) AddBatch(service string, messages []string) ([]int, error) {
	msg := map[string]interface{}{
		"service":  service,
		"messages": messages,
	}
	var resIDs []int
	err := c.apiCall("addBatch", msg, &resIDs)
	if err != nil {
		err = essentials.AddCtx("add log records", err)
	}
	return resIDs, err
}

// AddMedia adds a media record and returns its ID.
func (c *Client) AddMedia(folder, filename, mime string, data []byte, replace bool) (int, error) {
	msg := map[string]interface{}{
		"folder":   folder,
		"filename": filename,
		"mime":     mime,
		"data":     data,
		"replace":  replace,
	}
	var resID int
	err := c.apiCall("addMedia", msg, &resID)
	if err != nil {
		err = essentials.AddCtx("add media record", err)
	}
	return resID, err
}

// Overview returns the most recent log message from every
// service.
func (c *Client) Overview() ([]LogRecord, error) {
	msg := map[string]string{}
	var reply []LogRecord
	if err := c.apiCall("overview", msg, &reply); err != nil {
		return nil, essentials.AddCtx("fetch overview", err)
	}
	return reply, nil
}

// ServiceLog returns the log records for a service,
// sorted by most to least recent.
// It returns with an error if the service does not exist.
func (c *Client) ServiceLog(service string) ([]LogRecord, error) {
	msg := map[string]string{"service": service}
	var reply []LogRecord
	if err := c.apiCall("serviceLog", msg, &reply); err != nil {
		return nil, essentials.AddCtx("fetch service log", err)
	}
	return reply, nil
}

// Delete deletes the log for a service.
func (c *Client) Delete(service string) error {
	msg := map[string]string{"service": service}
	var result bool
	err := c.apiCall("delete", msg, &result)
	return essentials.AddCtx("delete service log", err)
}

// FullStream creates a channel of live log messages.
// The cancel chan can be closed to tell the stream to
// terminate.
// The returned channels will be closed on error or after
// a graceful shutdown.
func (c *Client) FullStream(cancel <-chan struct{}) (<-chan LogRecord, <-chan error) {
	return c.streamCall(cancel, "/api/fullStream", "")
}

// ServiceStream is like FullStream, but it limits
// messages to a specific service.
func (c *Client) ServiceStream(service string, cancel <-chan struct{}) (<-chan LogRecord,
	<-chan error) {
	escaped := url.QueryEscape(service)
	return c.streamCall(cancel, "/api/serviceStream", "service="+escaped)
}

func (c *Client) apiCall(name string, msg, reply interface{}) error {
	u := c.rootURL
	u.Path = "/api/" + name
	query, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	res, err := c.c.Post(u.String(), "application/json", bytes.NewReader(query))
	if res != nil {
		defer res.Body.Close()
	}
	if err != nil {
		return err
	}
	contents, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	var respObj struct {
		Data  interface{} `json:"data"`
		Error string      `json:"error"`
	}
	respObj.Data = reply
	if err := json.Unmarshal(contents, &respObj); err != nil {
		return errors.New(err.Error() + ": " + string(contents))
	}
	if respObj.Error != "" {
		return errors.New("remote error: " + respObj.Error)
	}
	if reply != nil {
		dataJSON, _ := json.Marshal(respObj.Data)
		if err := json.Unmarshal(dataJSON, reply); err != nil {
			return essentials.AddCtx("unmarshal data", err)
		}
	}
	return nil
}

func (c *Client) streamCall(done <-chan struct{}, path, query string) (<-chan LogRecord,
	<-chan error) {
	resChan := make(chan LogRecord, 1)
	errChan := make(chan error, 1)
	go func() {
		defer close(resChan)
		defer close(errChan)

		u := c.websocketURL()
		u.Path = path
		u.RawQuery = query

		conn, err := net.Dial("tcp", u.Host)
		if err != nil {
			errChan <- essentials.AddCtx("stream log", err)
			return
		}

		// Create dummy request for the AddCookie magic.
		req, err := http.NewRequest("GET", c.rootURL.String(), nil)
		if err != nil {
			errChan <- essentials.AddCtx("stream log", err)
			return
		}
		for _, c := range c.c.Jar.Cookies(&c.rootURL) {
			req.AddCookie(c)
		}
		req.Header.Set("Host", hostname(u.Host))

		cli, _, err := websocket.NewClient(conn, u, req.Header, 100, 100)
		if err != nil {
			errChan <- essentials.AddCtx("stream log", err)
			return
		}

		cleanupChan := make(chan struct{})
		defer close(cleanupChan)
		go func() {
			select {
			case <-done:
			case <-cleanupChan:
			}
			cli.Close()
		}()

		for {
			var msg LogRecord
			err := cli.ReadJSON(&msg)
			select {
			case <-done:
				return
			default:
			}
			if err != nil {
				errChan <- essentials.AddCtx("stream log", err)
				return
			}
			select {
			case resChan <- msg:
			case <-done:
				return
			}
		}
	}()
	return resChan, errChan
}

func (c *Client) websocketURL() *url.URL {
	u := c.rootURL
	if m, _ := regexp.MatchString(":[0-9]*$", u.Host); !m {
		if u.Scheme == "http" {
			u.Host += ":80"
		} else if u.Scheme == "https" {
			u.Host += ":443"
		}
	}
	if u.Scheme == "http" {
		u.Scheme = "ws"
	} else if u.Scheme == "https" {
		u.Scheme = "wss"
	}
	return &u
}

func hostname(h string) string {
	expr := regexp.MustCompile(":[0-9]*$")
	return expr.ReplaceAllString(h, "")
}
