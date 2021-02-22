package math

import "math"

func MinUI(x uint, y uint) uint {
	if x < y {
		return x
	}
	return y
}

func RadToDeg(v float64) float64 {
	return v * (180 / math.Pi)
}
