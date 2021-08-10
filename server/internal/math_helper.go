package internal

import (
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
