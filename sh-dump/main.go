// Command sh-dump prints the entire backlog for a given
// service.
package main

import (
	"fmt"
	"os"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/statushub"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage: sh-dump <service>")
		fmt.Fprintln(os.Stderr, "")
		statushub.PrintEnvUsage(os.Stderr)
		os.Exit(1)
	}

	client, err := statushub.AuthCLI()
	if err != nil {
		essentials.Die(err)
	}

	log, err := client.ServiceLog(os.Args[1])
	if err != nil {
		essentials.Die(err)
	}

	for i := len(log) - 1; i >= 0; i-- {
		fmt.Println(log[i].Message)
	}
}
