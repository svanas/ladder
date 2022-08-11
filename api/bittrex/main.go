package bittrex

import (
	"strings"
	"time"
)

const (
	appId      = "214"
	apiBase    = "https://api.bittrex.com"
	apiVersion = "v3"
)

var (
	cooldown    bool
	lastRequest time.Time
)

const (
	intensityLow   = 1  // 1 req/second
	intensityTwo   = 2  // 0.5 req/second
	intensitySuper = 60 // 1 req/minute
)

func requestsPerSecond(intensity int) float64 {
	return float64(1) / float64(intensity)
}

type call struct {
	path      string
	intensity int
}

var calls = []call{}

func getRequestsPerSecond(path string) (float64, bool) { // -> (rps, cooldown)
	if cooldown {
		cooldown = false
		return requestsPerSecond(intensitySuper), true
	}
	for i := range path {
		if strings.Contains("?", string(path[i])) {
			path = path[:i]
			break
		}
	}
	for _, call := range calls {
		if call.path == path {
			return requestsPerSecond(call.intensity), false
		}
	}
	return requestsPerSecond(intensityLow), false
}

func beforeRequest(path string) (bool, error) { // -> (cooled, error)
	elapsed := time.Since(lastRequest)
	rps, cooled := getRequestsPerSecond(path)
	if elapsed.Seconds() < (float64(1) / rps) {
		time.Sleep(time.Duration((float64(time.Second) / rps)) - elapsed)
	}
	return cooled, nil
}

func afterRequest() {
	lastRequest = time.Now()
}

func handleRateLimitErr(path string, cooled bool) {
	var (
		exists bool
	)
	for idx := range path {
		if strings.Contains("?", string(path[idx])) {
			path = path[:idx]
			break
		}
	}
	for idx := range calls {
		if calls[idx].path == path {
			if cooled {
				// rate limited immediately after a cooldown?
				// 1. do another round of "cooling down"
				// 2. do not slow this endpoint down just yet.
			} else {
				calls[idx].intensity = calls[idx].intensity + 1
			}
			exists = true
		}
	}
	if !exists {
		calls = append(calls, call{
			path,
			intensityTwo,
		})
	}
	cooldown = true
}
