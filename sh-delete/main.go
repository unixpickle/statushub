package main

import (
	"os"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/statushub"
)

func main() {
	if len(os.Args) != 2 {
		essentials.Die("Usage: sh-delete <service>")
	}

	client, err := statushub.AuthCLI()
	if err != nil {
		essentials.Die("Failed to create client:", err)
	}

	essentials.Must(client.Delete(os.Args[1]))
}
