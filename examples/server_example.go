package main

import (
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly"
)

func main() {
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
