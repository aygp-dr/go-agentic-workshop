package benchmarks

import (
	"math/rand"
	"time"
)

var rnd *rand.Rand

func init() {
	// Use a properly seeded random source instead of deprecated rand.Seed
	source := rand.NewSource(time.Now().UnixNano())
	rnd = rand.New(source)
}

// Random returns a random float64 in [0, 1)
func Random() float64 {
	return rnd.Float64()
}
