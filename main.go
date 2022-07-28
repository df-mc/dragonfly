package main

import (
	"fmt"
	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/cube/trace"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/chat"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/pelletier/go-toml"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"math"
	"os"
	"time"
)

func main() {
	log := logrus.New()
	log.Formatter = &logrus.TextFormatter{ForceColors: true}
	log.Level = logrus.DebugLevel

	chat.Global.Subscribe(chat.StdoutSubscriber{})

	config, err := readConfig()
	if err != nil {
		log.Fatalln(err)
	}

	srv := server.New(&config, log)
	srv.CloseOnProgramEnd()
	if err := srv.Start(); err != nil {
		log.Fatalln(err)
	}

	for srv.Accept(func(p *player.Player) {
		p.Handle(newRedstonePlayerHandler(p))
	}) {
	}
}

// readConfig reads the configuration from the config.toml file, or creates the file if it does not yet exist.
func readConfig() (server.Config, error) {
	c := server.DefaultConfig()
	if _, err := os.Stat("config.toml"); os.IsNotExist(err) {
		data, err := toml.Marshal(c)
		if err != nil {
			return c, fmt.Errorf("failed encoding default config: %v", err)
		}
		if err := ioutil.WriteFile("config.toml", data, 0644); err != nil {
			return c, fmt.Errorf("failed creating config: %v", err)
		}
		return c, nil
	}
	data, err := ioutil.ReadFile("config.toml")
	if err != nil {
		return c, fmt.Errorf("error reading config: %v", err)
	}
	if err := toml.Unmarshal(data, &c); err != nil {
		return c, fmt.Errorf("error decoding config: %v", err)
	}
	return c, nil
}

type redstonePlayerHandler struct {
	player.NopHandler
	p         *player.Player
	closeChan chan struct{}
}

func newRedstonePlayerHandler(p *player.Player) *redstonePlayerHandler {
	h := &redstonePlayerHandler{
		p:         p,
		closeChan: make(chan struct{}, 1),
	}
	go h.tick()
	return h
}

func (h *redstonePlayerHandler) tick() {
	t := time.NewTicker(time.Second / 20)
	for {
		select {
		case <-h.closeChan:
			return
		case <-t.C:
			yaw, pitch := h.p.Rotation()
			yaw, pitch = yaw*(math.Pi/180), pitch*(math.Pi/180)
			y := -math.Sin(pitch)
			xz := math.Cos(pitch)
			x := -xz * math.Sin(yaw)
			z := xz * math.Cos(yaw)
			start := h.p.Position().Add(mgl64.Vec3{0, h.p.EyeHeight()})
			direction := mgl64.Vec3{x, y, z}.Normalize().Mul(50)
			var hitBlock world.Block
			trace.TraverseBlocks(start, start.Add(direction), func(pos cube.Pos) bool {
				b := h.p.World().Block(pos)
				if _, ok := b.(block.Air); !ok {
					hitBlock = b
					return false
				}
				return true
			})
			if hitBlock != nil {
				popup := fmt.Sprintf("%T", hitBlock)
				switch hitBlock := hitBlock.(type) {
				case block.RedstoneDust:
					popup += fmt.Sprintf("\nPower: %d", hitBlock.Power)
				}
				h.p.SendPopup(popup)
			} else {
				h.p.SendPopup("You are not looking at a block")
			}
		}
	}
}

func (h *redstonePlayerHandler) HandleQuit() {
	close(h.closeChan)
}
