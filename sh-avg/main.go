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
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/statushub"
)

var DefaultAvgSizes = []int{10, 20, 50}

func main() {
	flags := ParseFlags()

	client, err := statushub.AuthCLI()
	if err != nil {
		essentials.Die(err)
	}

	serviceNames, err := ServiceNames(client, flags.ServiceName)
	essentials.Must(err)

	for {
		err := ProduceAggregates(client, flags, serviceNames)
		if flags.LoopInterval == 0 {
			essentials.Must(err)
			break
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		time.Sleep(flags.LoopInterval)
	}
}

// ProduceAggregates computes aggregates for the services
// and outputs the results to the appropriate places.
func ProduceAggregates(c *statushub.Client, f *Flags, serviceNames []string) error {
	printLine := func(message string) {
		fmt.Println(message)
		if f.LogName != "" {
			if _, err := c.Add(f.LogName, message); err != nil {
				fmt.Fprintln(os.Stderr, "Failed to log:", err)
			}
		}
	}
	for _, name := range serviceNames {
		if len(serviceNames) > 1 {
			printLine("Service: " + name)
		}
		log, err := c.ServiceLog(name)
		if err != nil {
			return err
		}

		fields := ExtractFields(log)
		if f.AvgSize == 0 {
			for _, size := range DefaultAvgSizes {
				printLine(AggSummary(size, fields, f.AggMethod))
			}
		} else {
			printLine(AggSummary(f.AvgSize, fields, f.AggMethod))
		}
	}
	return nil
}

// ExtractFields finds fields of the form "key=value" in a
// list of log messages.
//
// Returns a map from field names to a full history of the
// values for that field.
func ExtractFields(log []statushub.LogRecord) map[string][]float64 {
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

// ServiceNames gets service names matching an expression.
func ServiceNames(c *statushub.Client, expr string) ([]string, error) {
	if expr != "*" {
		return []string{expr}, nil
	}
	var serviceNames []string
	overview, err := c.Overview()
	if err != nil {
		return nil, err
	}
	for _, x := range overview {
		serviceNames = append(serviceNames, x.Service)
	}
	return serviceNames, nil
}
