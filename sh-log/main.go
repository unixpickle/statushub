package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"sync"

	"github.com/unixpickle/statushub"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: sh-log <service> [cmd [args...]]")
		fmt.Fprintln(os.Stderr, "")
		statushub.PrintEnvUsage(os.Stderr)
		os.Exit(1)
	}

	client, err := statushub.AuthCLI()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to create client:", err)
		os.Exit(1)
	}

	if len(os.Args) == 2 {
		logAndEcho(client, os.Stdin, os.Stdout)
	} else {
		logCommand(client, os.Args[2], os.Args[3:]...)
	}
}

func logCommand(c *statushub.Client, name string, args ...string) {
	cmd := exec.Command(name, args...)
	pipe1, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to create stdout pipe:", err)
		os.Exit(1)
	}
	pipe2, err := cmd.StderrPipe()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to create stderr pipe:", err)
		os.Exit(1)
	}

	var wg sync.WaitGroup
	outs := []io.Writer{os.Stdout, os.Stderr}
	for i, pipe := range []io.Reader{pipe1, pipe2} {
		wg.Add(1)
		go func(pipe io.Reader, out io.Writer) {
			defer wg.Done()
			logAndEcho(c, pipe, out)
		}(pipe, outs[i])
	}

	cmd.Start()

	// Forward our signals so the child can do graceful
	// shutdown if it wants to.
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		for _ = range c {
		}
	}()

	wg.Wait()
	cmd.Wait()
}

func logAndEcho(c *statushub.Client, in io.Reader, echo io.Writer) {
	r := bufio.NewReader(in)
	for {
		line, err := r.ReadString('\n')
		if len(line) == 0 && err != nil {
			return
		}
		if line[len(line)-1] == '\n' {
			line = line[:len(line)-1]
		}
		if _, err := c.Add(os.Args[1], line); err != nil {
			fmt.Fprintln(os.Stderr, "Failed to log:", err)
		}
		fmt.Fprintln(echo, line)
	}
}
