package binance

import (
	"context"
	"time"

	"github.com/adshao/go-binance/v2"
)

var (
	lastRequest       time.Time
	lastWeight        int     = 1
	requestsPerSecond float64 = 0
)

func getRequestsPerSecondFromInfo(info binance.ExchangeInfo) int64 {
	getIntervalNum := func(rl binance.RateLimit) int64 {
		if rl.IntervalNum > 0 {
			return rl.IntervalNum
		}
		return 1
	}

	for _, rl := range info.RateLimits {
		if rl.RateLimitType == "REQUEST_WEIGHT" {
			if rl.Interval == "SECOND" {
				return rl.Limit / getIntervalNum(rl)
			}
			if rl.Interval == "MINUTE" {
				return (rl.Limit / getIntervalNum(rl)) / 60
			}
			if rl.Interval == "DAY" {
				return (rl.Limit / getIntervalNum(rl)) / (24 * 60 * 60)
			}
		}
	}

	return 20
}

func getRequestsPerSecondFromClient(client binance.Client, weight int) float64 {
	var out float64 = 20

	if requestsPerSecond == 0 {
		info, err := client.NewExchangeInfoService().Do(context.Background())
		if err == nil {
			requestsPerSecond = float64(getRequestsPerSecondFromInfo(*info))
		}
	}

	if requestsPerSecond > 0 {
		out = requestsPerSecond
	}

	if lastWeight > 1 {
		out = out / float64(lastWeight)
	}
	lastWeight = weight

	return out
}

func beforeRequest(client binance.Client, request request) {
	elapsed := time.Since(lastRequest)
	rps := getRequestsPerSecondFromClient(client, weight[request])
	if elapsed.Seconds() < (float64(1) / rps) {
		time.Sleep(time.Duration((float64(time.Second) / rps)) - elapsed)
	}
}

func afterRequest() {
	lastRequest = time.Now()
}
