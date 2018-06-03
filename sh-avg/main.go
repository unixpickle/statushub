// Command sh-avg reads a StatusHub log and computes
// averages for various recurrent values in the log.
//
// In general, values are found by searching for tokens of
// the form "key=value", where key is alphabetical and
// value is numeric.
// For example, take a log with the following output:
//
//     2017/02/03 09:39:00 iter 23369: cost=2.9909554 validation=2.130182
//     2017/02/03 09:39:01 iter 23370: cost=2.5142508
//     2017/02/03 09:39:02 iter 23371: cost=2.1001065
//     2017/02/03 09:39:03 iter 23372: cost=1.3731229
//     2017/02/03 09:39:06 iter 23373: cost=1.9894226 validation=3.5907815
//     2017/02/03 09:39:06 iter 23374: cost=2.245893
//     2017/02/03 09:39:07 iter 23375: cost=2.1684818
//     2017/02/03 09:39:08 iter 23376: cost=3.9548137
//     2017/02/03 09:39:11 iter 23377: cost=1.9482157 validation=2.5914357
//
// From this, sh-avg would extract two keys: "cost" and
// "validation".
package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/statushub"
)

type AggFn func([]float64) float64

var AggMethods = map[string]AggFn{
	"mean":   computeMean,
	"median": computeMedian,
	"min":    computeMin,
	"max":    computeMax,
}

func main() {
	var aggregateType string
	flag.StringVar(&aggregateType, "type", "mean",
		"the type of aggregate (mean, median, max, min)")
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

	aggregateMethod, ok := AggMethods[aggregateType]
	if !ok {
		essentials.Die("unknown aggregate type:", aggregateType)
	}

	var avgSize int
	if len(flag.Args()) == 2 {
		var err error
		avgSize, err = strconv.Atoi(flag.Args()[1])
		if err != nil {
			essentials.Die("invalid average size:", os.Args[2])
		}
	}

	client, err := statushub.AuthCLI()
	if err != nil {
		essentials.Die(err)
	}

	var serviceNames []string
	if flag.Args()[0] != "*" {
		serviceNames = []string{flag.Args()[0]}
	} else {
		overview, err := client.Overview()
		if err != nil {
			essentials.Die(err)
		}
		for _, x := range overview {
			serviceNames = append(serviceNames, x.Service)
		}
	}

	for _, name := range serviceNames {
		if len(serviceNames) > 1 {
			fmt.Println("Service:", name)
		}
		log, err := client.ServiceLog(name)
		if err != nil {
			essentials.Die(err)
		}

		fields := getFields(log)
		if avgSize == 0 {
			for _, size := range []int{10, 20, 50} {
				printAggregates(size, fields, aggregateMethod)
			}
		} else {
			printAggregates(avgSize, fields, aggregateMethod)
		}
	}
}

func getFields(log []statushub.LogRecord) map[string][]float64 {
	exp := regexp.MustCompile(`^([a-zA-Z_0-9\-]*)=([0-9\.\-e]*)$`)
	res := map[string][]float64{}
	for _, record := range log {
		for _, field := range strings.Fields(record.Message) {
			m := exp.FindStringSubmatch(field)
			if m == nil {
				continue
			}
			fieldName := m[1]
			fieldVal := m[2]
			valFloat, err := strconv.ParseFloat(fieldVal, 64)
			if err == nil {
				res[fieldName] = append(res[fieldName], valFloat)
			}
		}
	}
	return res
}

func printAggregates(size int, fields map[string][]float64, aggFn AggFn) {
	aggs := map[string]float64{}
	fieldNames := []string{}
	for key, vals := range fields {
		fieldNames = append(fieldNames, key)
		aggs[key] = aggFn(vals)
	}
	sort.Strings(fieldNames)
	fmt.Printf("size %d:", size)
	for _, name := range fieldNames {
		fmt.Printf(" %s=%f", name, aggs[name])
	}
	fmt.Println()
}

func computeMean(values []float64) float64 {
	sum := 0.0
	for _, val := range values {
		sum += val
	}
	return sum / float64(len(values))
}

func computeMedian(values []float64) float64 {
	sort.Float64s(values)
	if len(values)%2 != 0 {
		return values[len(values)/2]
	} else {
		return (values[len(values)/2-1] + values[len(values)/2]) / 2
	}
}

func computeMin(values []float64) float64 {
	res := values[0]
	for _, val := range values[1:] {
		res = math.Min(res, val)
	}
	return res
}

func computeMax(values []float64) float64 {
	res := values[0]
	for _, val := range values[1:] {
		res = math.Max(res, val)
	}
	return res
}
