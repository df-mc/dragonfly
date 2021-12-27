package main

import (
	"fmt"
	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/player/chat"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/pelletier/go-toml"
	"github.com/sirupsen/logrus"
	"io/ioutil"
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

	for {
		if p, err := srv.Accept(); err != nil {
			return
		} else {
			p.SetGameMode(world.GameModeSurvival)
			go func() {
				overworld := p.World()
				nether, end := overworld.PortalDestinations()
				time.Sleep(5 * time.Second)
				if !p.Dead() {
					nether.AddEntity(p)
					p.Teleport(nether.Spawn().Vec3Centre())
					time.Sleep(5 * time.Second)
					if !p.Dead() {
						end.AddEntity(p)
						p.Teleport(end.Spawn().Vec3Centre())
						time.Sleep(5 * time.Second)
						if !p.Dead() {
							overworld.AddEntity(p)
							p.Teleport(overworld.Spawn().Vec3Centre())
						}
					}
				}
			}()
		}
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
