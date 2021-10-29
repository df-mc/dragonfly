package main

import (
	"fmt"
	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/chat"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/pelletier/go-toml"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

import _ "net/http/pprof"

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
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
	srv.World().StartWeatherCycle()
	srv.World().StartThundering(time.Hour)

	for {
		if p, err := srv.Accept(); err != nil {
			return
		} else {
			p.Inventory().AddItem(item.NewStack(item.Snowball{}, 128))
			p.Handle(h{p: p})
		}
	}
}

type h struct {
	player.NopHandler
	p *player.Player
}

func (h h) HandleItemUse(ctx *event.Context) {
	v := entity.DirectionVector(h.p)
	pos := h.p.Position().Add(mgl64.Vec3{0, h.p.EyeHeight()})
	for (h.p.World().Block(cube.PosFromVec3(pos)) == block.Air{}) {
		pos = pos.Add(v)
	}
	h.p.World().AddEntity(entity.NewLightning(pos.Add(mgl64.Vec3{0, 1})))
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
