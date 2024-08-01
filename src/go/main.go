package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"runtime"
	"runtime/pprof"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

const (
	READ_SIZE      = 4096 * 8192
	EOL       byte = '\n'
	SEMICOLON byte = ';'
	DASH      byte = '-'
	DOT       byte = '.'
	ZERO      byte = '0'
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

func (r *Result) addTemp(temp int16) {
	if temp < r.min {
		r.min = temp
	} else if temp > r.max {
		r.max = temp
	}
	r.sum += int64(temp)
	r.count++
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
	fi, err := file.Stat()
	if err != nil {
		return err
	}
	size := fi.Size()
	filedata, err := syscall.Mmap(int(file.Fd()), 0, int(size), syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		log.Fatalf("Mmap: %v", err)
	}

	defer func() {
		if err := syscall.Munmap(filedata); err != nil {
			log.Fatalf("Munmap: %v", err)
		}
	}()

	// control # of threads base on cores
	cores := runtime.NumCPU()
	semaphore := make(chan struct{}, cores)
	chunkChannel := make(chan map[string]*Result, cores)

	// start listening to channel before we start pushing to it
	aggChannel := make(chan map[string]*Result, 1)
	defer close(aggChannel)
	go aggregateResults(chunkChannel, aggChannel)

	var wg sync.WaitGroup

	data := filedata
	for {
		// make slice a little bigger than what we actually read so we can read to next end of line without reallocation
		var n int = len(data)
		if n > READ_SIZE {
			n = READ_SIZE
		}
		//chunk := data[:]
		// find EOL
		for n < len(data) {
			if data[n] != EOL {
				n++
			} else {
				break
			}
		}
		chunk := data[:n]
		n++
		if n >= len(data) {
			break
		}
		data = data[n:]

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

func processChunk(chunk []byte) map[string]*Result {
	results := make(map[string]*Result, 10000)

	for lineIdx, chunkSize := 0, len(chunk); lineIdx < chunkSize; {

		nameStart := lineIdx
		nameIdx := lineIdx
		for ; nameIdx < chunkSize && chunk[nameIdx] != SEMICOLON; nameIdx++ {

		}
		station := string(chunk[nameStart:nameIdx])
		tempStart := nameIdx + 1
		tempIdx := tempStart
		for ; tempIdx < chunkSize && chunk[tempIdx] != EOL; tempIdx++ {

		}
		temp := parseNumber(chunk[tempStart:tempIdx])
		lineIdx = tempIdx + 1
		result, ok := results[station]
		if ok {
			result.addTemp(temp)
		} else {
			results[station] = NewResult(temp)
		}
	}

	// fmt.Printf("chunk results size %d\n", len(results))
	return results
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

// assumes the slice is just the right length
func parseNumber(chunk []byte) int16 {
	tempIdx := 0
	var isNegative = false
	if chunk[tempIdx] == DASH {
		isNegative = true
		tempIdx++
	}

	sum := int16(0)
	if chunk[tempIdx+1] == DOT {
		// single digit number
		sum = int16((chunk[tempIdx] - ZERO)) * 10
		tempIdx += 2
	} else {
		// 2 digit number
		sum = int16((chunk[tempIdx]-ZERO))*100 + int16((chunk[tempIdx+1])-ZERO)*10
		tempIdx += 3
	}
	sum += int16(chunk[tempIdx] - ZERO)
	if isNegative {
		sum = -sum
	}
	return sum
}
