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
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

// A LogRecord is one logged message.
type LogRecord struct {
	Service string `json:"serviceName"`
	Message string `json:"message"`
	Time    int64  `json:"time"`
	ID      int    `json:"id"`
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
		return nil, errors.New("create cookie jar: " + err.Error())
	}
	u, err := url.Parse(rootURL)
	if err != nil {
		return nil, errors.New("bad root URL: " + err.Error())
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
		err = errors.New("add log record: " + err.Error())
	}
	return resID, err
}

// Overview returns the most recent log message from every
// service.
func (c *Client) Overview() ([]LogRecord, error) {
	msg := map[string]string{}
	var reply []LogRecord
	if err := c.apiCall("overview", msg, &reply); err != nil {
		return nil, errors.New("fetch overview: " + err.Error())
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
		return nil, errors.New("fetch service log: " + err.Error())
	}
	return reply, nil
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
			return errors.New("unmarshal data: " + err.Error())
		}
	}
	return nil
}
