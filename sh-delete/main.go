package main

import (
	"fmt"
	"os"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/statushub"
)

func main() {
	if len(os.Args) < 2 {
		essentials.Die("Usage: sh-delete <service> [service ...]")
	}

	client, err := statushub.AuthCLI()
	if err != nil {
		essentials.Die("Failed to create client:", err)
	}

	for _, service := range os.Args[1:] {
		if err := client.Delete(service); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
}
