package server

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player/chat"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

type Console struct {
	Subscriber *chat.Chat
}

func (Console) Name() string         { return "Console" }
func (Console) Position() mgl64.Vec3 { return mgl64.Vec3{0, 0, 0} }
func (Console) World() *world.World  { return nil }
func (console Console) SendCommandOutput(output *cmd.Output) {
	for _, e := range output.Errors() {
		fmt.Fprintf(console.Subscriber, "error: %v", e)
	}
	for _, m := range output.Messages() {
		fmt.Fprint(console.Subscriber, m)
	}
}
func (console Console) ReadInput() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		if scanner.Scan() {
			commandName := strings.Split(scanner.Text(), " ")[0]
			commandArgs := strings.TrimPrefix(scanner.Text(), commandName+" ")
			if command, ok := cmd.ByAlias(commandName); ok {
				command.Execute(commandArgs, console)
			} else {
				fmt.Fprintf(console.Subscriber, "Unknown command: %s", commandName)
			}
		}
	}
}
