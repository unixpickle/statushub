package main

import (
	"fmt"
	"math"
	"sort"
)

type AggFn func([]float64) float64

var AggMethods = map[string]AggFn{
	"mean":   computeMean,
	"median": computeMedian,
	"min":    computeMin,
	"max":    computeMax,
}

// AggSummary produces a string that summarizes the fields
// from some set of log messages.
func AggSummary(size int, fields map[string][]float64, aggFn AggFn) string {
	aggs := map[string]float64{}
	fieldNames := []string{}
	for key, vals := range fields {
		fieldNames = append(fieldNames, key)
		if len(vals) > size {
			vals = vals[:size]
		}
		aggs[key] = aggFn(vals)
	}
	sort.Strings(fieldNames)
	res := fmt.Sprintf("size %d:", size)
	for _, name := range fieldNames {
		res += fmt.Sprintf(" %s=%f", name, aggs[name])
	}
	return res
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
