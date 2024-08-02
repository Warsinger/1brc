package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"runtime"
	"runtime/pprof"
	"sort"
	"strconv"
	"strings"
	"sync"
)

const (
	READ_SIZE      = 32 * 1024 * 1024
	EOL       byte = '\n'
	SEMICOLON byte = ';'
	DASH      byte = '-'
	DOT       byte = '.'
	ZERO      byte = '0'
	ZEROD2         = '0' * (10 + 1)
	ZEROD3         = '0' * (100 + 10 + 1)
)

type Result struct {
	min   int16
	max   int16
	sum   int64
	count int
}

func NewResult(temp int16) *Result {
	return &Result{temp, temp, int64(temp), 1}
}

func (result *Result) accumulateResult(other *Result) {
	if other.min < result.min {
		result.min = other.min
	} else if other.max > result.max {
		result.max = other.max
	}
	result.sum += other.sum
	result.count += other.count
}

func (r *Result) String() string {
	return fmt.Sprintf("%.1f/%.1f/%.1f", round(float64(r.min)/10.0), r.average(), round(float64(r.max)/10.0)) //, r.sum, r.count)
}

func (r Result) average() float64 {
	// return round(float64(r.sum/r.count) / 10.0)
	return round(float64(r.sum) / 10.0 / float64(r.count))
}

func round(x float64) float64 {
	return roundJava(x*10.0) / 10.0
}

// roundJava returns the closest integer to the argument, with ties
// rounding to positive infinity, see java's Math.round
func roundJava(x float64) float64 {
	t := math.Trunc(x)
	if x < 0.0 && t-x == 0.5 {
		//return t
	} else if math.Abs(x-t) >= 0.5 {
		t += math.Copysign(1, x)
	}

	if t == 0 { // check -0
		return 0.0
	}
	return t
}

func main() {
	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to file")
	fileName := flag.String("file", "", "path to measurements file")
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	if *fileName == "" {
		log.Fatal("Filename not specified")
	}

	file, err := os.Open(*fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	err = processFile(file)
	if err != nil {
		log.Fatal(err)
	}
}

func processFile(file *os.File) error {
	r := bufio.NewReaderSize(file, READ_SIZE)
	// control # of threads base on cores
	cores := runtime.NumCPU()
	semaphore := make(chan struct{}, cores)
	chunkChannel := make(chan map[string]*Result, cores)

	// start listening to channel before we start pushing to it
	aggChannel := make(chan map[string]*Result, 1)
	defer close(aggChannel)
	go aggregateResults(chunkChannel, aggChannel)

	var wg sync.WaitGroup

	for {
		// make slice a little bigger than what we actually read so we can read to next end of line without reallocation
		chunk := make([]byte, READ_SIZE, READ_SIZE+128)

		n, err := r.Read(chunk)
		chunk = chunk[:n]

		if n == 0 {
			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Println(err)
				break
			}
			return err
			// } else {
			// 	fmt.Printf("read %d bytes\n", n)
		}
		nextUntillNewline, err := r.ReadBytes(EOL)
		// fmt.Printf("%d bytes till new line\n", len(nextUntillNewline))
		if err != io.EOF {
			chunk = append(chunk, nextUntillNewline...)
		}

		semaphore <- struct{}{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			chunkChannel <- processChunk(chunk)
			<-semaphore
		}()
	}
	// fmt.Println("all reading done")
	wg.Wait()
	close(chunkChannel)
	close(semaphore)
	// fmt.Println("all processing done")
	results := <-aggChannel

	printResults(results)

	return nil
}

func aggregateResults(chunkChannel, aggChannel chan map[string]*Result) {
	results := make(map[string]*Result, 100000)
	for chunkResults := range chunkChannel {
		// fmt.Printf("agg chunk %d of size %d\n", count, len(chunkResults))
		for station, chunkResult := range chunkResults {
			result, ok := results[station]
			if !ok {
				results[station] = chunkResult
				// fmt.Printf("new result %v %v\n", temp, result)
			} else {
				result.accumulateResult(chunkResult)
				// fmt.Printf("existing result %v %v\n", temp, result)
			}
		}
	}

	aggChannel <- results
}

// custom hashing to speed things up copied from https://github.com/gunnarmorling/1brc/blob/4daeb94b048e074c2b80aac1074b68eb92285ea8/src/main/go/AlexanderYastrebov/calc.go#L132
func processChunk(data []byte) map[string]*Result {
	// Use fixed size linear probe lookup table
	const (
		// use power of 2 for fast modulo calculation,
		// should be larger than max number of keys which is 10_000
		entriesSize = 1 << 14

		// use FNV-1a hash
		fnv1aOffset64 = 14695981039346656037
		fnv1aPrime64  = 1099511628211
	)

	type entry struct {
		m     Result
		hash  uint64
		vlen  int
		value [128]byte // use power of 2 > 100 for alignment
	}
	entries := make([]entry, entriesSize)
	entriesCount := 0

	// keep short and inlinable
	getResult := func(hash uint64, value []byte) *Result {
		i := hash & uint64(entriesSize-1)
		entry := &entries[i]

		// bytes.Equal could be commented to speedup assuming no hash collisions
		for entry.vlen > 0 && !(entry.hash == hash) { // && bytes.Equal(entry.value[:entry.vlen], value)) {
			i = (i + 1) & uint64(entriesSize-1)
			entry = &entries[i]
		}

		if entry.vlen == 0 {
			entry.hash = hash
			entry.vlen = copy(entry.value[:], value)
			entriesCount++
		}
		return &entry.m
	}

	// assume valid input
	for len(data) > 0 {

		idHash := uint64(fnv1aOffset64)
		semiPos := 0
		for i, b := range data {
			if b == SEMICOLON {
				semiPos = i
				break
			}

			// calculate FNV-1a hash
			idHash ^= uint64(b)
			idHash *= fnv1aPrime64
		}

		idData := data[:semiPos]

		data = data[semiPos+1:]

		var temp int16
		// parseNumber
		{
			negative := data[0] == DASH
			if negative {
				data = data[1:]
			}

			_ = data[3]
			if data[1] == DOT {
				// 1.2\n
				temp = int16(data[0])*10 + int16(data[2]) - ZEROD2
				data = data[4:]
				// 12.3\n
			} else {
				_ = data[4]
				temp = int16(data[0])*100 + int16(data[1])*10 + int16(data[3]) - ZEROD3
				data = data[5:]
			}

			if negative {
				temp = -temp
			}
		}

		m := getResult(idHash, idData)
		if m.count == 0 {
			m.min = temp
			m.max = temp
			m.sum = int64(temp)
			m.count = 1
		} else {
			m.min = min(m.min, temp)
			m.max = max(m.max, temp)
			m.sum += int64(temp)
			m.count++
		}
	}

	result := make(map[string]*Result, entriesCount)
	for i := range entries {
		entry := &entries[i]
		if entry.m.count > 0 {
			result[string(entry.value[:entry.vlen])] = &entry.m
		}
	}
	return result
}

func printResults(results map[string]*Result) {
	keys := make([]string, 0, len(results))

	for k := range results {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		fmt.Printf("%s=%s\n", k, results[k])
	}
}

func parseTemp(value string) (int, error) {
	tempStr := strings.Split(value, ".")
	return strconv.Atoi(tempStr[0] + tempStr[1])
}
