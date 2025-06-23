package benchmarks

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Random returns a random float64 in [0, 1)
func Random() float64 {
	return rand.Float64()
}