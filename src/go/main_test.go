package main

import (
	"fmt"
	"testing"
)

func TestRound(t *testing.T) {
	for _, tc := range []struct {
		value    float64
		expected string
	}{
		{value: -1.5, expected: "-1.0"},
		{value: -1.0, expected: "-1.0"},
		{value: -0.7, expected: "-1.0"},
		{value: -0.5, expected: "0.0"},
		{value: -0.3, expected: "0.0"},
		{value: 0.0, expected: "0.0"},
		{value: 0.3, expected: "0.0"},
		{value: 0.5, expected: "1.0"},
		{value: 0.7, expected: "1.0"},
		{value: 1.0, expected: "1.0"},
		{value: 1.5, expected: "2.0"},
		{value: -31.0500, expected: "-31.0"},
	} {
		if rounded := roundJava(tc.value); fmt.Sprintf("%.1f", rounded) != tc.expected {
			t.Errorf("Wrong rounding of %v, expected: %s, got: %.1f", tc.value, tc.expected, rounded)
		}
	}
}
func TestAverage(t *testing.T) {
	for _, tc := range []struct {
		value    Result
		expected string
	}{
		{value: Result{sum: -1863, count: 6}, expected: "-31.0"},
		{value: Result{sum: 2524, count: 14}, expected: "18.0"},
		{value: Result{sum: -241, count: 2}, expected: "-12.0"},
		{value: Result{sum: 1870, count: 4}, expected: "46.8"},
		{value: Result{sum: 44, count: 3}, expected: "1.5"},
	} {
		if avg := tc.value.average(); fmt.Sprintf("%.1f", avg) != tc.expected {
			t.Errorf("Wrong average of %v, expected: %s, got: %.1f", tc.value, tc.expected, avg)
		}
	}
}

func TestParseNumber(t *testing.T) {
	for _, tc := range []struct {
		value    string
		expected int16
	}{
		{"-99.9", -999}, {"-12.3", -123}, {"-1.5", -15}, {"-1.0", -10},
		{"0.0", 0}, {"0.3", 3}, {"12.3", 123}, {"99.9", 999},
	} {
		t.Run(tc.value, func(t *testing.T) {
			results := processChunk([]byte("Station;" + tc.value + "\n"))
			result := results["Station"]
			if len(results) != 1 || result == nil {
				t.Fatalf("unexpected stations: %v", results)
			}
			if result.min != tc.expected || result.max != tc.expected || result.sum != int64(tc.expected) || result.count != 1 {
				t.Fatalf("parsing %s: got min=%d max=%d sum=%d count=%d, want temperature %d and count 1", tc.value, result.min, result.max, result.sum, result.count, tc.expected)
			}
		})
	}
}
