package bitstamp

import (
	"time"
)

const (
	endpoint = "https://www.bitstamp.net/api/v2"
)

var (
	lastRequest       time.Time
	requestsPerSecond float64 = 10
)

func beforeRequest(method, path string) error {
	elapsed := time.Since(lastRequest)
	if elapsed.Seconds() < (float64(1) / requestsPerSecond) {
		time.Sleep(time.Duration((float64(time.Second) / requestsPerSecond) - float64(elapsed)))
	}
	return nil
}

func afterRequest() {
	lastRequest = time.Now()
}
