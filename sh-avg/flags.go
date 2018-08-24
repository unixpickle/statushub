package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/statushub"
)

type Flags struct {
	AggMethod    AggFn
	AvgSize      int
	ServiceName  string
	LoopInterval time.Duration
	LogName      string
}

func ParseFlags() *Flags {
	f := &Flags{}

	var aggregateType string
	flag.StringVar(&aggregateType, "type", "mean", "the type of aggregate (mean, median, max, min)")
	flag.DurationVar(&f.LoopInterval, "loop", 0, "interval for computing averages in a loop")
	flag.StringVar(&f.LogName, "log", "", "StatusHub service to which results are echoed")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: sh-avg [flags] <service|*> [avg size]")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Flags:")
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, "")
		statushub.PrintEnvUsage(os.Stderr)
	}
	flag.Parse()

	if len(flag.Args()) != 1 && len(flag.Args()) != 2 {
		flag.Usage()
		os.Exit(1)
	}

	f.ServiceName = flag.Args()[0]

	var ok bool
	f.AggMethod, ok = AggMethods[aggregateType]
	if !ok {
		essentials.Die("unknown aggregate type:", aggregateType)
	}

	if len(flag.Args()) == 2 {
		var err error
		f.AvgSize, err = strconv.Atoi(flag.Args()[1])
		if err != nil {
			essentials.Die("invalid average size:", flag.Args()[1])
		}
	}

	return f
}
