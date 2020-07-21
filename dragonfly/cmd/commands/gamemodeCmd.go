package commands

import (
	"github.com/df-mc/dragonfly/dragonfly"
	"github.com/df-mc/dragonfly/dragonfly/cmd"
	"github.com/df-mc/dragonfly/dragonfly/player"
	"github.com/df-mc/dragonfly/dragonfly/world/gamemode"
	"strings"
)

//
// basic implementation of vanilla gamemode command
//
type GamemodeCommand struct {
	Mode   string
	Player string `optional:""`
}

func (args GamemodeCommand) Run(source cmd.Source, output *cmd.Output) {
	sender := source.(*player.Player)
	if args.Mode == "" || args.Mode == "/gamemode" {
		sender.Message("Usage: /gamemode <mode> <player>")
		return
	}

	// check if the Mode argument is a valid gamemode to switch to
	mode, m := parseGamemode(args.Mode)
	if mode == nil {
		sender.Message(m)
		return
	}

	// checks if the player parameter is present or the player argument is the same as the sender
	if args.Player == "" || strings.ToLower(sender.Name()) == strings.ToLower(args.Player) {
		sender.SetGameMode(mode)
		sender.Message("Set own gamemode to " + m)
		return
	}

	// checks is the player argument is a player currently online
	server := dragonfly.Server{}
	target, _ := server.GetPlayer(args.Player)
	if target == nil {
		sender.Message("That player cannot be found")
		return
	}

	// finally, changes the players gamemode based on the mode parsed
	target.SetGameMode(mode)
	target.Message("Your gamemode has been updated to " + m)
	sender.Message("You set " + target.Name() + "'s gamemode to " + m)
}

func parseGamemode(mode string) (gamemode.GameMode, string) {
	if mode == "0" || mode == "s" {
		return gamemode.Survival{}, "Survival mode"
	}
	if mode == "1" || mode == "c" {
		return gamemode.Creative{}, "Creative mode"
	}
	if mode == "2" || mode == "a" {
		return gamemode.Adventure{}, "Adventure mode"
	}
	if mode == "3" || mode == "" {
		return gamemode.Spectator{}, "Spectator mode"
	}
	return nil, "Unknown gamemode"
}
