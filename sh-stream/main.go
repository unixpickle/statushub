package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/statushub"
)

func main() {
	var n int
	var reconnect bool
	flag.IntVar(&n, "n", 0, "max number of messages")
	flag.BoolVar(&reconnect, "reconnect", false, "automatically attempt reconnect")
	flag.Parse()

	if len(flag.Args()) != 0 && len(flag.Args()) != 1 {
		fmt.Fprintln(os.Stderr, "Usage: sh-stream [flags ...] [service]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	for {
		// Re-create client each time to avoid having
		// the session expire.
		client, err := statushub.AuthCLI()
		if err != nil {
			essentials.Die(err)
		}
		if err := stream(client, n); err != nil {
			fmt.Fprintln(os.Stderr, err)
			if !reconnect {
				os.Exit(1)
			}
		}
	}
}

func stream(client *statushub.Client, n int) error {
	var stream <-chan statushub.LogRecord
	var errChan <-chan error

	if len(flag.Args()) == 0 {
		stream, errChan = client.FullStream(nil)
	} else {
		stream, errChan = client.ServiceStream(flag.Args()[0], nil)
	}

	for i := 0; i < n || n == 0; i++ {
		message, ok := <-stream
		if !ok {
			break
		}
		fmt.Println(message.Message)
	}

	return <-errChan
}
