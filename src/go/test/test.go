package main

import (
	"fmt"
	"math"
)

// There's some weird stuff going on here with rounding based on whether the data is in the struct or not
func main1() {
	// var sum int = -1863
	// var count int = 6
	var sum int = -1534
	var count int = 4
	var avg float64

	r := Result{sum: sum, count: count}
	avg = r.average()
	fmt.Printf("avg struct:%.2f\n", avg)

	avg = average1(float64(sum), float64(count))
	fmt.Printf("avg plain1 %.2f\n", avg)

	avg = average2(sum, count)
	fmt.Printf("avg plain2 %.2f\n", avg)

	b := make([]byte, 50)
	for i := 0; i < len(b); i++ {
		b[i] = byte(i) + 45
	}
	bx := b[4:10]
	fmt.Printf("b %v\n", string(b))
	fmt.Printf("bx %v %v %v\n", cap(bx), len(bx), bx)
}

func (r *Result) average() float64 {
	return average2(r.sum, r.count)
}

func average2(sum, count int) float64 {
	return average1(float64(sum), float64(count))
}

func average1(sum, count float64) float64 {
	return round(sum / 10.0 / count)
}

type Result struct {
	sum   int
	count int
}

func round(x float64) float64 {
	return math.Floor((x+0.05)*10) / 10
}

func round1(x float64) float64 {
	return roundJava(x*10.0) / 10.0
}

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

func (r *Result) String() string {
	return fmt.Sprintf("%.2f{%d/%d}", r.average(), r.sum, r.count)
}
