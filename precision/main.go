package precision

import (
	"fmt"
	"strconv"
	"strings"
)

func Parse(value string) int {
	i := strings.Index(value, ".")
	if i > -1 {
		n := i + 1
		for n < len(value) {
			if value[n] != '0' {
				return n - i
			}
			n++
		}
	}
	return 0
}

// Round sets the number of places after the decimal
func Round(x float64, prec int) float64 {
	out, _ := strconv.ParseFloat(fmt.Sprintf("%.[2]*[1]f", x, prec), 64)
	return out
}
