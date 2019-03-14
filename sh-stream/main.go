package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/statushub"
)

func main() {
	var n int
	var reconnect bool
	var timeout time.Duration
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: sh-stream [flags] [service]")
		flag.PrintDefaults()
	}
	flag.IntVar(&n, "n", 0, "max number of messages")
	flag.BoolVar(&reconnect, "reconnect", false, "automatically attempt reconnect")
	flag.DurationVar(&timeout, "timeout", 0, "max time between log messages")
	flag.Parse()

	if len(flag.Args()) != 0 && len(flag.Args()) != 1 {
		flag.Usage()
		os.Exit(1)
	}

	for {
		// Re-create client each time to avoid having
		// the session expire.
		client, err := statushub.AuthCLI()
		if err != nil {
			essentials.Die(err)
		}
		if err := stream(client, n, timeout); err != nil {
			fmt.Fprintln(os.Stderr, err)
			if !reconnect {
				os.Exit(1)
			}
		}
	}
}

func stream(client *statushub.Client, n int, timeout time.Duration) error {
	var stream <-chan statushub.LogRecord
	var errChan <-chan error

	if len(flag.Args()) == 0 {
		stream, errChan = client.FullStream(nil)
	} else {
		stream, errChan = client.ServiceStream(flag.Args()[0], nil)
	}

	var timer *time.Timer
	var timerCh <-chan time.Time
	if timeout != 0 {
		timer = time.NewTimer(timeout)
		timerCh = timer.C
	}
	for i := 0; i < n || n == 0; i++ {
		select {
		case message, ok := <-stream:
			if !ok {
				break
			}
			fmt.Println(message.Message)
			if timer != nil {
				if !timer.Stop() {
					<-timer.C
				}
				timer.Reset(timeout)
			}
		case <-timerCh:
			essentials.Die("timeout expired")
		}

	}

	return <-errChan
}
