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
		expected string
	}{
		{value: "-99.9", expected: "-999"},
		{value: "-12.3", expected: "-123"},
		{value: "-1.5", expected: "-15"},
		{value: "-1.0", expected: "-10"},
		{value: "0.0", expected: "0"},
		{value: "0.3", expected: "3"},
		{value: "12.3", expected: "123"},
		{value: "99.9", expected: "999"},
	} {
		if number, _ := parseTemp(tc.value); fmt.Sprintf("%d", number) != tc.expected {
			t.Errorf("Wrong parsing of %v, expected: %s, got: %d", tc.value, tc.expected, number)
		}
	}
}
