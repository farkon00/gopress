package shared

import "math"

func Ceil_div(x, y int) int {
	return int(math.Ceil(float64(x) / float64(y)))
}
