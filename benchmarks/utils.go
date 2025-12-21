package benchmarks

import (
	"math/rand/v2"
)

// Random returns a random float64 in [0, 1)
func Random() float64 {
	return rand.Float64()
}
