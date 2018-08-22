package main

import (
	"io"
	"regexp"
	"time"

	"github.com/unixpickle/essentials"
)

const LogTimeFormat = "2006/01/02 15:04:05"

// A Message is a log message as it goes through the log
// pre-processing pipeline.
type Message struct {
	Line string

	// Dest is the local file to which the message should
	// eventually be echoed.
	Dest io.Writer

	// Filtered is set to true if the log message should
	// not be sent to the server.
	Filtered bool
}

// Pipeline creates a message processing pipeline based on
// the command-line arguments.
func Pipeline(f *Flags) (chan<- *Message, <-chan *Message) {
	input := make(chan *Message, 1)
	var output <-chan *Message = input
	if f.AddTimestamps {
		output = AddTimestamps(output, f.Timezone)
	}
	output = Filter(output, f.Filter)
	output = TakeInterval(output, f.LineInterval)
	output = Buffer(output, f.Buffer)
	return input, output
}

// ReadBuffer reads at least one message and up to the
// entire buffer of messages.
func ReadBuffer(msgs <-chan *Message) []*Message {
	msg, ok := <-msgs
	if !ok {
		return nil
	}
	res := []*Message{msg}
	for i := 0; i < cap(msgs)-1; i++ {
		select {
		case newMsg, ok := <-msgs:
			if !ok {
				return res
			}
			res = append(res, newMsg)
		default:
			return res
		}
	}
	return res
}

// AddTimestamps is a pipeline stage that adds timestamps
// to the beginning of every log message.
func AddTimestamps(messages <-chan *Message, timezone string) <-chan *Message {
	var location *time.Location
	if timezone != "" {
		var err error
		location, err = time.LoadLocation(timezone)
		if err != nil {
			essentials.Die("Invalid timezone:", err)
		}
	}
	return pipelineStage(messages, func(msg *Message) {
		t := time.Now()
		if location != nil {
			t = t.In(location)
		}
		msg.Line = t.Format(LogTimeFormat) + " " + msg.Line
	})
}

// Filter is a pipeline stage that filters log messages
// for a given regular expression.
func Filter(messages <-chan *Message, filter string) <-chan *Message {
	if filter == "" {
		return messages
	}
	expr, err := regexp.Compile(filter)
	essentials.Must(err)
	return pipelineStage(messages, func(msg *Message) {
		msg.Filtered = msg.Filtered || !expr.MatchString(msg.Line)
	})
}

// TakeInterval is a pipeline stage that filters every log
// message except for every n-th one.
func TakeInterval(messages <-chan *Message, interval int) <-chan *Message {
	var lineIndex int
	return pipelineStage(messages, func(msg *Message) {
		if !msg.Filtered {
			msg.Filtered = (lineIndex%interval != 0)
			lineIndex++
		}
	})
}

// Buffer adds a write buffer to the pipeline.
func Buffer(messages <-chan *Message, bufSize int) <-chan *Message {
	res := make(chan *Message, bufSize)
	go func() {
		defer close(res)
		for msg := range messages {
			res <- msg
		}
	}()
	return res
}

func pipelineStage(messages <-chan *Message, f func(*Message)) <-chan *Message {
	res := make(chan *Message, 1)
	go func() {
		defer close(res)
		for msg := range messages {
			f(msg)
			res <- msg
		}
	}()
	return res
}
