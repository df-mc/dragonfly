package main

import (
	"fmt"
	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/cube/trace"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/chat"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/pelletier/go-toml"
	"github.com/sirupsen/logrus"
	"os"
	"time"
)

func main() {
	log := logrus.New()
	log.Formatter = &logrus.TextFormatter{ForceColors: true}
	log.Level = logrus.DebugLevel

	chat.Global.Subscribe(chat.StdoutSubscriber{})

	conf, err := readConfig(log)
	if err != nil {
		log.Fatalln(err)
	}

	srv := conf.New()
	srv.CloseOnProgramEnd()

	srv.Listen()
	for srv.Accept(func(p *player.Player) {
		p.Handle(newRedstonePlayerHandler(p))
	}) {
	}
}

// readConfig reads the configuration from the config.toml file, or creates the
// file if it does not yet exist.
func readConfig(log server.Logger) (server.Config, error) {
	c := server.DefaultConfig()
	var zero server.Config
	if _, err := os.Stat("config.toml"); os.IsNotExist(err) {
		data, err := toml.Marshal(c)
		if err != nil {
			return zero, fmt.Errorf("encode default config: %v", err)
		}
		if err := os.WriteFile("config.toml", data, 0644); err != nil {
			return zero, fmt.Errorf("create default config: %v", err)
		}
		return c.Config(log)
	}
	data, err := os.ReadFile("config.toml")
	if err != nil {
		return zero, fmt.Errorf("read config: %v", err)
	}
	if err := toml.Unmarshal(data, &c); err != nil {
		return zero, fmt.Errorf("decode config: %v", err)
	}
	return c.Config(log)
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
	p.ShowCoordinates()
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
			w := h.p.World()
			start := h.p.Position().Add(mgl64.Vec3{0, h.p.EyeHeight()})
			end := start.Add(entity.DirectionVector(h.p).Mul(50))
			var hitBlock world.Block
			var hitPos cube.Pos
			trace.TraverseBlocks(start, end, func(pos cube.Pos) bool {
				b := w.Block(pos)
				if _, ok := b.(block.Air); !ok {
					hitBlock = b
					hitPos = pos
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
				popup += fmt.Sprintf("\nCalculated Power: %d", w.ReceivedRedstonePower(hitPos))
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
