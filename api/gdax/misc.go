package gdax

import (
	"strconv"
)

func ParseFloat(value string) float64 {
	out, err := strconv.ParseFloat(value, 64)
	if err == nil {
		return out
	}
	return 0
}
