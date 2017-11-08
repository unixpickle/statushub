package statushub

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/howeyc/gopass"
)

const (
	RootEnvVar = "STATUSHUB_ROOT"
	PassEnvVar = "STATUSHUB_PASS"
)

// AuthCLI uses environment variables (and potentially
// user input) to obtain an authenticated Client.
func AuthCLI() (*Client, error) {
	rootURL := os.Getenv(RootEnvVar)
	if rootURL == "" {
		return nil, errors.New("authenticate: missing " + RootEnvVar +
			" environment variable")
	}
	client, err := NewClient(rootURL)
	if err != nil {
		return nil, errors.New("authenticate: " + err.Error())
	}
	pass := os.Getenv(PassEnvVar)
	if pass == "" {
		fmt.Print("StatusHub password: ")
		passBytes, err := gopass.GetPasswd()
		if err != nil {
			return nil, err
		}
		pass = string(passBytes)
	}
	if err := client.Login(pass); err != nil {
		return nil, errors.New("authenticate: " + err.Error())
	}
	return client, nil
}

// PrintEnvUsage prints usage messages about the
// authentication environment variables.
func PrintEnvUsage(w io.Writer) error {
	messages := []string{
		"Set the " + RootEnvVar + " environment variable",
		"to the URL of the StatusHub server",
		"(e.g. http://localhost:8080).",
		"",
		"Set the " + PassEnvVar + " environment variable",
		"to the StatusHub password to avoid manual",
		"entry.",
	}
	for _, msg := range messages {
		if _, err := fmt.Fprintln(w, msg); err != nil {
			return errors.New("print environment usage: " + err.Error())
		}
	}
	return nil
}
