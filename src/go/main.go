package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Result struct {
	min   int16
	max   int16
	sum   int64
	count uint64
}

func NewResult(temp int16) *Result {
	return &Result{temp, temp, int64(temp), 1}
}

// func (result *Result) aggregate(other *Result) {
// 	if other.min < result.min {
// 		result.min = other.min
// 	} else if other.max > result.min {
// 		result.max = other.max
// 	}
// 	result.sum += other.sum
// 	result.count += other.count
// }

func (r *Result) addTemp(temp int16) {
	if temp < r.min {
		r.min = temp
	} else if temp > r.max {
		r.max = temp
	}
	r.sum += int64(temp)
	r.count++
}
func (r Result) String() string {
	return fmt.Sprintf("%.1f/%.1f/%.1f", round10(float64(r.min)), round(float64(r.sum)/10.0/float64(r.count)), round10(float64(r.max)))
}

func round(value float64) float64 {
	return math.Round(value*10.0) / 10.0
}
func round10(value float64) float64 {
	return math.Round(value) / 10.0
}

func main() {
	fileName := os.Args[1]
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	results := make(map[string]*Result, 10)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		name, temp := processLine(scanner.Text())
		aggregateLine(results, name, temp)
	}

	printResults(results)
}

func printResults(results map[string]*Result) {

	keys := make([]string, 0, len(results))

	for k := range results {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		fmt.Printf("%s=%s\n", k, *results[k])
	}
}

func aggregateLine(results map[string]*Result, name string, temp int16) {
	result, ok := results[name]
	if !ok {
		result = NewResult(temp)
		results[name] = result
		// fmt.Printf("new result %v %v\n", temp, result)
	} else {
		result.addTemp(temp)
		// fmt.Printf("existing result %v %v\n", temp, result)
	}
}

func processLine(line string) (string, int16) {
	// fmt.Println(line)
	parts := strings.Split(line, ";")
	tempStr := strings.Split(parts[1], ".")
	temp, err := strconv.Atoi(tempStr[0] + tempStr[1])
	if err != nil {
		log.Fatal(err)
	}
	return parts[0], int16(temp)
}
