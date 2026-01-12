package common

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConvertTimeToInterval(t *testing.T) {
	t.Parallel()

	expectedResults := map[time.Duration]string{
		0 * time.Millisecond: "0ms-5ms",
		1 * time.Millisecond: "0ms-5ms",
		5 * time.Millisecond: "0ms-5ms",

		6 * time.Millisecond:  "5ms-10ms",
		9 * time.Millisecond:  "5ms-10ms",
		10 * time.Millisecond: "5ms-10ms",

		11 * time.Millisecond: "10ms-20ms",
		19 * time.Millisecond: "10ms-20ms",
		20 * time.Millisecond: "10ms-20ms",

		21 * time.Millisecond: "20ms-35ms",
		34 * time.Millisecond: "20ms-35ms",
		35 * time.Millisecond: "20ms-35ms",

		36 * time.Millisecond: "35ms-50ms",
		49 * time.Millisecond: "35ms-50ms",
		50 * time.Millisecond: "35ms-50ms",

		51 * time.Millisecond: "50ms-75ms",
		74 * time.Millisecond: "50ms-75ms",
		75 * time.Millisecond: "50ms-75ms",

		76 * time.Millisecond:  "75ms-100ms",
		99 * time.Millisecond:  "75ms-100ms",
		100 * time.Millisecond: "75ms-100ms",

		101 * time.Millisecond: "100ms-125ms",
		124 * time.Millisecond: "100ms-125ms",
		125 * time.Millisecond: "100ms-125ms",

		126 * time.Millisecond: "125ms-150ms",
		149 * time.Millisecond: "125ms-150ms",
		150 * time.Millisecond: "125ms-150ms",

		151 * time.Millisecond: "150ms-175ms",
		174 * time.Millisecond: "150ms-175ms",
		175 * time.Millisecond: "150ms-175ms",

		176 * time.Millisecond: "175ms-200ms",
		199 * time.Millisecond: "175ms-200ms",
		200 * time.Millisecond: "175ms-200ms",

		201 * time.Millisecond: "200ms-300ms",
		299 * time.Millisecond: "200ms-300ms",
		300 * time.Millisecond: "200ms-300ms",

		301 * time.Millisecond: "300ms-400ms",
		399 * time.Millisecond: "300ms-400ms",
		400 * time.Millisecond: "300ms-400ms",

		401 * time.Millisecond: "400ms-500ms",
		499 * time.Millisecond: "400ms-500ms",
		500 * time.Millisecond: "400ms-500ms",

		501 * time.Millisecond: "500ms-750ms",
		749 * time.Millisecond: "500ms-750ms",
		750 * time.Millisecond: "500ms-750ms",

		752 * time.Millisecond:  "750ms-1s",
		999 * time.Millisecond:  "750ms-1s",
		1000 * time.Millisecond: "750ms-1s",

		1001 * time.Millisecond: "1s-2s",
		1999 * time.Millisecond: "1s-2s",
		2000 * time.Millisecond: "1s-2s",

		2001 * time.Millisecond: "2s-5s",
		4999 * time.Millisecond: "2s-5s",
		5000 * time.Millisecond: "2s-5s",

		5001 * time.Millisecond: ">5s",
		10 * time.Second:        ">5s",
		50 * time.Second:        ">5s",
	}

	for key, value := range expectedResults {
		assert.Equal(t, value, ConvertTimeToInterval(key))
	}
}

func TestGetAllPerformanceIntervals(t *testing.T) {
	t.Parallel()

	expectedIntervals := []string{
		"0ms-5ms",
		"5ms-10ms",
		"10ms-20ms",
		"20ms-35ms",
		"35ms-50ms",
		"50ms-75ms",
		"75ms-100ms",
		"100ms-125ms",
		"125ms-150ms",
		"150ms-175ms",
		"175ms-200ms",
		"200ms-300ms",
		"300ms-400ms",
		"400ms-500ms",
		"500ms-750ms",
		"750ms-1s",
		"1s-2s",
		"2s-5s",
		">5s",
	}

	assert.Equal(t, expectedIntervals, GetAllPerformanceIntervals())
}
