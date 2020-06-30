package main

import (
	"github.com/df-mc/dragonfly/dragonfly"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	server := dragonfly.New(nil, nil)
	server.CloseOnProgramEnd()
	if err := server.Start(); err != nil {
		panic(err)
	}
	for {
		if _, err := server.Accept(); err != nil {
			return
		}
	}
}
