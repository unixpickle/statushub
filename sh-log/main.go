package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"sync"

	"github.com/creack/pty"
	"github.com/unixpickle/essentials"
	"github.com/unixpickle/statushub"
)

func main() {
	flags, args := ParseFlags()

	client, err := statushub.AuthCLI()
	if err != nil {
		essentials.Die("Failed to create client:", err)
	}

	if flags.Clear {
		client.Delete(flags.ServiceName)
	}

	pipelineIn, pipelineOut := Pipeline(flags)

	go func() {
		defer close(pipelineIn)
		if len(args) == 0 {
			linesToMessages(pipelineIn, os.Stdin, os.Stdout)
		} else {
			logCommand(pipelineIn, args[0], args[1:]...)
		}
	}()

	submitMessages(client, flags, pipelineOut)
}

func submitMessages(c *statushub.Client, f *Flags, messages <-chan *Message) {
	for {
		msgs := ReadBuffer(messages)
		if len(msgs) == 0 {
			return
		}
		if _, err := c.AddBatch(f.ServiceName, unfilteredMessages(msgs)); err != nil {
			fmt.Fprintln(os.Stderr, "Failed to log:", err)
		}
		for _, msg := range msgs {
			fmt.Fprintln(msg.Dest, msg.Line)
		}
	}
}

func unfilteredMessages(msgs []*Message) []string {
	res := []string{}
	for _, msg := range msgs {
		if !msg.Filtered {
			res = append(res, msg.Line)
		}
	}
	return res
}

func logCommand(msgCh chan<- *Message, name string, args ...string) {
	cmd := exec.Command(name, args...)

	ptmx, err := pty.Start(cmd)
	if err != nil {
		essentials.Die("Failed to start command:", err)
	}

	defer ptmx.Close()
	if err := disableEcho(ptmx); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to disable stdin echo:", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func(pipe io.Reader, out io.Writer) {
		defer wg.Done()
		linesToMessages(msgCh, pipe, out)
	}(ptmx, os.Stdout)

	go func() {
		io.Copy(ptmx, os.Stdin)
	}()

	// Catch our first Ctrl+C so the child can do graceful
	// shutdown if it wants to.
	//
	// If the child logs a ton of stuff on exit, then the
	// user can press Ctrl+C again to terminate sh-log before
	// all the output has been sent to the server.
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		cmd.Process.Signal(os.Interrupt)
		signal.Stop(c)
	}()

	wg.Wait()
	cmd.Wait()
}

func linesToMessages(msgCh chan<- *Message, in io.Reader, echo io.Writer) {
	r := bufio.NewReader(in)
	for {
		line, err := r.ReadString('\n')
		if len(line) == 0 && err != nil {
			return
		}
		if line[len(line)-1] == '\n' {
			line = line[:len(line)-1]
		}
		msgCh <- &Message{
			Line: line,
			Dest: echo,
		}
	}
}
