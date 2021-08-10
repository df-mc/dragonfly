package internal

import (
	"math"
	"math/rand"
	"time"
)

func NextIntn(max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max)
}

func NextInt() int {
	rand.Seed(time.Now().UnixNano())
	return rand.Int()
}

func Floor(a float64) int {
	return int(math.Floor(a))
}
