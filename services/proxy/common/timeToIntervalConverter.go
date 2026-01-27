package common

import (
	"fmt"
	"time"
)

var intervals = []time.Duration{
	5 * time.Millisecond,
	10 * time.Millisecond,
	20 * time.Millisecond,
	35 * time.Millisecond,
	50 * time.Millisecond,
	75 * time.Millisecond,
	100 * time.Millisecond,
	125 * time.Millisecond,
	150 * time.Millisecond,
	175 * time.Millisecond,
	200 * time.Millisecond,
	300 * time.Millisecond,
	400 * time.Millisecond,
	500 * time.Millisecond,
	750 * time.Millisecond,
	1 * time.Second,
	2 * time.Second,
	5 * time.Second,
}

func ConvertTimeToInterval(timeToWait time.Duration) string {
	for i := 0; i < len(intervals); i++ {
		if intervals[i] < timeToWait {
			continue
		}

		return convertAtCounter(i)
	}

	return convertOverTheMaximum()
}

func convertAtCounter(counter int) string {
	if counter == 0 {
		return makeInterval("0ms", intervals[0].String())
	}

	return makeInterval(intervals[counter-1].String(), intervals[counter].String())
}

func convertOverTheMaximum() string {
	return fmt.Sprintf(">%s", intervals[len(intervals)-1])
}

func makeInterval(from, to string) string {
	return fmt.Sprintf("%s-%s", from, to)
}

// GetAllPerformanceIntervals will output all the intervals defined
func GetAllPerformanceIntervals() []string {
	ret := make([]string, 0, len(intervals)+1)
	for i := 0; i < len(intervals); i++ {
		ret = append(ret, convertAtCounter(i))
	}
	ret = append(ret, convertOverTheMaximum())
	return ret
}
