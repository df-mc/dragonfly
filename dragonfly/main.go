package main

import (
	"bytes"
	"fmt"
	"github.com/df-mc/dragonfly/dragonfly"
	"github.com/df-mc/dragonfly/dragonfly/player/chat"
	"github.com/pelletier/go-toml"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
)

func main() {
	log := logrus.New()
	log.Formatter = &logrus.TextFormatter{ForceColors: true}
	log.Level = logrus.DebugLevel

	chat.Global.Subscribe(chat.StdoutSubscriber{})

	if !loopbackExempted() {
		const loopbackExemptCmd = `CheckNetIsolation LoopbackExempt -a -n="Microsoft.MinecraftUWP_8wekyb3d8bbwe"`
		log.Printf("You are currently unable to join the server on this machine. Run %v in an admin PowerShell session to be able to.\n", loopbackExemptCmd)
	}

	config, err := readConfig()
	if err != nil {
		log.Fatalln(err)
	}

	server := dragonfly.New(&config, log)
	server.CloseOnProgramEnd()
	if err := server.Start(); err != nil {
		log.Fatalln(err)
	}

	for {
		if _, err := server.Accept(); err != nil {
			return
		}
	}
}

// loopbackExempted checks if the user has has loopback enabled
// The user will need this in order to connect to their server locally.
func loopbackExempted() bool {
	if runtime.GOOS != "windows" {
		return true
	}
	data, _ := exec.Command("CheckNetIsolation", "LoopbackExempt", "-s", `-n="microsoft.minecraftuwp_8wekyb3d8bbwe"`).CombinedOutput()
	return bytes.Contains(data, []byte("microsoft.minecraftuwp_8wekyb3d8bbwe"))
}

// readConfig reads the configuration from the config.toml file, or creates the file if it does not yet exist.
func readConfig() (dragonfly.Config, error) {
	c := dragonfly.DefaultConfig()
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
