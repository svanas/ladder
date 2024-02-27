package oneinch

import (
	"time"
)

var (
	lastRequest       time.Time
	requestsPerSecond float64 = 1
)

func beforeRequest() {
	elapsed := time.Since(lastRequest)
	if elapsed.Seconds() < (float64(1) / requestsPerSecond) {
		time.Sleep(time.Duration((float64(time.Second) / requestsPerSecond)) - elapsed)
	}
}

func afterRequest() {
	lastRequest = time.Now()
}
