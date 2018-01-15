package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"time"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/statushub"
)

const LogTimeFormat = "2006/01/02 15:04:05"

type Flags struct {
	ServiceName   string
	AddTimestamps bool
	Timezone      string
}

func ParseFlags() (f *Flags, args []string) {
	f = &Flags{}

	flag.BoolVar(&f.AddTimestamps, "timestamps", false, "prepend timestamps to lines")
	flag.StringVar(&f.Timezone, "timezone", "", "show timestamps in an IANA timezone")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: sh-log [flags] <service> [cmd [args...]]")
		fmt.Fprintln(os.Stderr, "")
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, "")
		statushub.PrintEnvUsage(os.Stderr)
	}

	flag.Parse()
	if len(flag.Args()) == 0 {
		flag.Usage()
		os.Exit(1)
	}
	f.ServiceName = flag.Args()[0]

	return f, flag.Args()[1:]
}

func main() {
	flags, args := ParseFlags()

	client, err := statushub.AuthCLI()
	if err != nil {
		essentials.Die("Failed to create client:", err)
	}

	if len(args) == 0 {
		logAndEcho(client, flags, os.Stdin, os.Stdout)
	} else {
		logCommand(client, flags, args[0], args[1:]...)
	}
}

func logCommand(c *statushub.Client, f *Flags, name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	pipe1, err := cmd.StdoutPipe()
	if err != nil {
		essentials.Die("Failed to create stdout pipe:", err)
	}
	pipe2, err := cmd.StderrPipe()
	if err != nil {
		essentials.Die("Failed to create stderr pipe:", err)
	}

	var wg sync.WaitGroup
	outs := []io.Writer{os.Stdout, os.Stderr}
	for i, pipe := range []io.Reader{pipe1, pipe2} {
		wg.Add(1)
		go func(pipe io.Reader, out io.Writer) {
			defer wg.Done()
			logAndEcho(c, f, pipe, out)
		}(pipe, outs[i])
	}

	if err := cmd.Start(); err != nil {
		essentials.Die("Failed to start command:", err)
	}

	// Ignore our first Ctrl+C so the child can do graceful
	// shutdown if it wants to.
	//
	// If the child logs a ton of stuff on exit, then the
	// user can press Ctrl+C again to terminate sh-log before
	// all the output has been sent to the server.
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		signal.Stop(c)
	}()

	wg.Wait()
	cmd.Wait()
}

func logAndEcho(c *statushub.Client, f *Flags, in io.Reader, echo io.Writer) {
	r := bufio.NewReader(in)
	for {
		line, err := r.ReadString('\n')
		if len(line) == 0 && err != nil {
			return
		}
		if line[len(line)-1] == '\n' {
			line = line[:len(line)-1]
		}
		if f.AddTimestamps {
			line = addTimestamp(f.Timezone, line)
		}
		if _, err := c.Add(f.ServiceName, line); err != nil {
			fmt.Fprintln(os.Stderr, "Failed to log:", err)
		}
		fmt.Fprintln(echo, line)
	}
}

func addTimestamp(timezone, line string) string {
	t := time.Now()
	if timezone != "" {
		location, err := time.LoadLocation(timezone)
		if err != nil {
			essentials.Die("Invalid timezone:", err)
		}
		t = t.In(location)
	}
	return t.Format(LogTimeFormat) + " " + line
}
