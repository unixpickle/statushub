package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/unixpickle/statushub"
)

type Flags struct {
	ServiceName   string
	AddTimestamps bool
	Timezone      string
	LineInterval  int
	Filter        string
	Buffer        int
}

func ParseFlags() (f *Flags, args []string) {
	f = &Flags{}

	flag.BoolVar(&f.AddTimestamps, "timestamps", false, "prepend timestamps to lines")
	flag.StringVar(&f.Timezone, "timezone", "", "show timestamps in an IANA timezone")
	flag.IntVar(&f.LineInterval, "interval", 1, "interval at which to log lines")
	flag.StringVar(&f.Filter, "filter", "", "regular expression to filter for log messages")
	flag.IntVar(&f.Buffer, "buffer", 100, "log message buffer size")
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
