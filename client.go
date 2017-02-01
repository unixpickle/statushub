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
	var resID float64
	err := c.apiCall("add", msg, &resID)
	return int(resID), err
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
