package main

import (
	"fmt"
	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/entity/effect"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/chat"
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

	handler1 := Handler1{
		player.NopHandler{},
	}

	handler2 := Handler2{
		player.NopHandler{},
	}

	for srv.Accept(func(p *player.Player) {
		for i := 0; i < 100000; i++ {
			p.AddHandler(player.NopHandler{})
		}

		// Add handlers to incoming players.
		changeHandler1 := p.AddHandler(handler1)
		changeHandler2 := p.AddHandler(handler2)

		// Remove handler1 5 seconds after joining.
		go func() {
			time.Sleep(5 * time.Second)
			changeHandler1(nil)
		}()

		// Change handler2 to handler1 10 seconds after joining.
		go func() {
			time.Sleep(10 * time.Second)
			changeHandler2(handler1)
		}()
	}) {
	}
}

type Handler1 struct {
	player.NopHandler
}

func (h Handler1) HandleMove(e player.EventMove) {
	e.Player.Message(e.Player.Position())
}

type Handler2 struct {
	player.NopHandler
}

func (h Handler2) HandleJump(e player.EventJump) {
	e.Player.AddEffect(effect.New(effect.Blindness{}, 1, time.Second))
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
