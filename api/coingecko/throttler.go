package coingecko

import (
	"time"
)

var lastRequest time.Time

func beforeRequest() {
	elapsed := time.Since(lastRequest)
	rps := apiRequestsPerSecond()
	if elapsed.Seconds() < (float64(1) / rps) {
		time.Sleep(time.Duration((float64(time.Second) / rps)) - elapsed)
	}
}

func afterRequest() {
	lastRequest = time.Now()
}
