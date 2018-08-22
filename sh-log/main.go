package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"sync"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/statushub"
)

func main() {
	flags, args := ParseFlags()

	client, err := statushub.AuthCLI()
	if err != nil {
		essentials.Die("Failed to create client:", err)
	}

	pipelineIn, pipelineOut := Pipeline(flags)

	go func() {
		defer close(pipelineIn)
		if len(args) == 0 {
			linesToMessages(pipelineIn, flags, os.Stdin, os.Stdout)
		} else {
			logCommand(pipelineIn, flags, args[0], args[1:]...)
		}
	}()

	submitMessages(client, flags, pipelineOut)
}

func submitMessages(c *statushub.Client, f *Flags, messages <-chan *Message) {
	for {
		msg, ok := <-messages
		if !ok {
			return
		}
		msgs := []*Message{msg}
	UnbufferLoop:
		for i := 0; i < f.Buffer-1; i++ {
			select {
			case newMsg, ok := <-messages:
				if !ok {
					break UnbufferLoop
				}
				msgs = append(msgs, newMsg)
			default:
				break UnbufferLoop
			}
		}
		// TODO: add all unfiltered messages in batch.
		for _, msg := range msgs {
			if !msg.Filtered {
				if _, err := c.Add(f.ServiceName, msg.Line); err != nil {
					fmt.Fprintln(os.Stderr, "Failed to log:", err)
				}
			}
			fmt.Fprintln(msg.Dest, msg.Line)
		}
	}
}

func logCommand(msgCh chan<- *Message, f *Flags, name string, args ...string) {
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
			linesToMessages(msgCh, f, pipe, out)
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

func linesToMessages(msgCh chan<- *Message, f *Flags, in io.Reader, echo io.Writer) {
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
