package main

import (
	"git.jetbrains.space/dragonfly/dragonfly/dragonfly"
)

func main() {
	server := dragonfly.New(nil, nil)
	server.CloseOnProgramEnd()
	if err := server.Run(); err != nil {
		panic(err)
	}
	for {
		_, _ = server.Accept()
	}
}
