package main

import (
	"github.com/dragonfly-tech/dragonfly/dragonfly"
)

func main() {
	server, err := dragonfly.New(nil, nil)
	if err != nil {
		panic(err)
	}
	server.CloseOnProgramEnd()
	if err := server.Run(); err != nil {
		panic(err)
	}
	for {
		server.Accept()
	}
}
