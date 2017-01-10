package main

import (
	"fmt"
	"os"

	"github.com/howeyc/gopass"
	"github.com/unixpickle/statushub"
)

const (
	RootEnvVar = "STATUSHUB_ROOT"
	PassEnvVar = "STATUSHUB_PASS"
)

func main() {
	rootURL := os.Getenv(RootEnvVar)
	if len(os.Args) != 3 || rootURL == "" {
		fmt.Fprintln(os.Stderr, "Usage: sh-log <service> <msg>")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Set the "+RootEnvVar+" environment variable")
		fmt.Fprintln(os.Stderr, "to the URL of the StatusHub server")
		fmt.Fprintln(os.Stderr, "(e.g. http://localhost:8080).")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Set the "+PassEnvVar+" environment variable")
		fmt.Fprintln(os.Stderr, "to the StatusHub password to avoid manual")
		fmt.Fprintln(os.Stderr, "entry.")
		os.Exit(1)
	}

	client, err := statushub.NewClient(rootURL)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to create client:", err)
		os.Exit(1)
	}

	if err := authenticate(client); err != nil {
		fmt.Fprintln(os.Stderr, "Authentication failed:", err)
		os.Exit(1)
	}

	if id, err := client.Add(os.Args[1], os.Args[2]); err != nil {
		fmt.Fprintln(os.Stderr, "Log failed:", err)
		os.Exit(1)
	} else {
		fmt.Println("Entry created with ID", id)
	}
}

func authenticate(c *statushub.Client) error {
	pass := os.Getenv(PassEnvVar)
	if pass == "" {
		fmt.Print("Password: ")
		passBytes, err := gopass.GetPasswd()
		if err != nil {
			return err
		}
		pass = string(passBytes)
	}
	return c.Login(pass)
}
