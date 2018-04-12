package main

import (
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"

	"github.com/howeyc/gopass"
	"github.com/unixpickle/essentials"
)

// DefaultLogSize is the default capacity of the status
// backlog.
const DefaultLogSize = 1000

// DefaultMediaCache is the default soft-limit on the
// number of bytes to store for a single media item's
// backlog.
const DefaultMediaCache = 10000000

// Config manages the server settings.
// It automatically deals with concurrency issues, saving,
// loading, and prompting the user for new values.
type Config struct {
	cfg  *configData
	path string
	lock sync.RWMutex
}

// LoadConfig loads the configuration from a path or
// creates a new one.
func LoadConfig(path string) (*Config, error) {
	contents, err := ioutil.ReadFile(path)
	if os.IsNotExist(err) {
		fmt.Print("New password: ")
		pass, err := gopass.GetPasswd()
		if err != nil {
			return nil, essentials.AddCtx("read password", err)
		}
		res := &Config{
			cfg: &configData{
				PasswordHash: hashPassword(string(pass)),
				LogSize:      DefaultLogSize,
				MediaCache:   DefaultMediaCache,
			},
			path: path,
		}
		if err := res.save(); err != nil {
			return nil, err
		}
		return res, nil
	}
	res := &Config{path: path}
	if err := json.Unmarshal(contents, &res.cfg); err != nil {
		return nil, err
	}
	return res, nil
}

// CheckPass checks if the given password is correct.
func (c *Config) CheckPass(p string) bool {
	c.lock.RLock()
	res := c.cfg.PasswordHash == hashPassword(p)
	c.lock.RUnlock()
	return res
}

// SetPass updates the password.
func (c *Config) SetPass(p string) error {
	return c.alter(func() {
		c.cfg.PasswordHash = hashPassword(p)
	})
}

// LogSize returns the current log size setting.
func (c *Config) LogSize() int {
	c.lock.RLock()
	res := c.cfg.LogSize
	c.lock.RUnlock()
	return res
}

// SetLogSize updates the log size setting.
func (c *Config) SetLogSize(s int) error {
	return c.alter(func() {
		c.cfg.LogSize = s
	})
}

// MediaCache returns the current media cache size.
func (c *Config) MediaCache() int {
	c.lock.RLock()
	res := c.cfg.MediaCache
	c.lock.RUnlock()
	return res
}

// SetMediaCache sets the media cache size.
func (c *Config) SetMediaCache(s int) error {
	return c.alter(func() {
		c.cfg.MediaCache = s
	})
}

func (c *Config) alter(f func()) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	old := *c.cfg
	f()
	if err := c.save(); err != nil {
		c.cfg = &old
		return err
	}
	return nil
}

func (c *Config) save() error {
	data, err := json.Marshal(c.cfg)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(c.path, data, 0600)
}

type configData struct {
	PasswordHash string `json:"pass"`
	LogSize      int    `json:"log_size"`
	MediaCache   int    `json:"media_cache"`
}

func hashPassword(p string) string {
	sum := sha512.Sum512([]byte(p))
	return strings.ToLower(hex.EncodeToString(sum[:]))
}
