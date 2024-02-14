package precision

import (
	"fmt"
	"math/big"
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

func S2F(s string) (*big.Float, error) {
	f, _, err := new(big.Float).Parse(s, 0)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func F2S(f big.Float, prec int) string {
	return f.Text('f', prec)
}
