package main

import (
	"github.com/df-mc/dragonfly/dragonfly"
	"github.com/df-mc/dragonfly/dragonfly/block"
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/world/gamemode"
)

func main() {
	server := dragonfly.New(nil, nil)
	server.CloseOnProgramEnd()
	if err := server.Start(); err != nil {
		panic(err)
	}
	for {
		if p, err := server.Accept(); err != nil {
			return
		} else {
			p.SetGameMode(gamemode.Creative{})
			p.Inventory().AddItem(item.NewStack(block.Gravel{}, 64))
			p.Inventory().AddItem(item.NewStack(block.Sand{}, 64))
			p.Inventory().AddItem(item.NewStack(block.Sandstone{}, 64))
			p.Inventory().AddItem(item.NewStack(block.SandstoneCut{}, 64))
			p.Inventory().AddItem(item.NewStack(block.SandstoneChiseled{}, 64))
			p.Inventory().AddItem(item.NewStack(item.Flint{}, 64))
		}
	}
}
