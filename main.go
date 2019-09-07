package main

import (
	"bytes"
	"dragonfly/dragonfly"
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
	if !loopbackExempted() {
		const loopbackExemptCmd = `CheckNetIsolation LoopbackExempt -a -n="Microsoft.MinecraftUWP_8wekyb3d8bbwe"`
		log.Printf("You are currently unable to join the server on this machine. Run %v in an admin PowerShell session to be able to.\n", loopbackExemptCmd)
	}

	config := readConfig(log)

	server := dragonfly.New(&config, log)
	if err := server.Run(); err != nil {
		log.Fatalln(err)
	}
}

// loopbackExempted checks if the user has has loopback enabled
// The user will need this in order to connect to their server locally.
func loopbackExempted() bool {
	if runtime.GOOS != "windows" {
		return true
	}
	data, _ := exec.Command("CheckNetIsolation", "LoopbackExempt", "-s", `-n="microsoft.minecraftuwp_8wekyb3d8bbwe"`).CombinedOutput()
	if bytes.Contains(data, []byte("microsoft.minecraftuwp_8wekyb3d8bbwe")) {
		return true
	}
	return false
}

// readConfig reads the configuration from the config.toml file, or creates the file if it does not yet exist.
func readConfig(log *logrus.Logger) dragonfly.Config {
	c := dragonfly.DefaultConfig()
	if _, err := os.Stat("config.toml"); os.IsNotExist(err) {
		data, err := toml.Marshal(c)
		if err != nil {
			log.Fatalf("Error encoding default config: %v\n", err)
		}
		if err := ioutil.WriteFile("config.toml", data, 0644); err != nil {
			log.Fatalf("Error creating config: %v\n", err)
		}
		return c
	}
	data, err := ioutil.ReadFile("config.toml")
	if err != nil {
		log.Fatalf("error reading config: %v", err)
	}
	if err := toml.Unmarshal(data, &c); err != nil {
		log.Fatalf("error decoding config: %v", err)
	}
	return c
}
