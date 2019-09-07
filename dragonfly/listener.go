package dragonfly

import (
	"bytes"
	"dragonfly/dragonfly/logger"
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/naoina/toml"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"os"
	"os/exec"
	"runtime"
)

type ServerConfig struct {
	ServerName string
	WorldName string
	MaxPlayers int
	Address string
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

const ascii = `
________                                   _____________        
___  __ \____________ _______ ________________  __/__  /____  __
__  / / /_  ___/  __ /_  __  /  __ \_  __ \_  /_ __  /__  / / /
_  /_/ /_  /   / /_/ /_  /_/ // /_/ /  / / /  __/ _  / _  /_/ / 
/_____/ /_/    \__,_/ _\__, / \____//_/ /_//_/    /_/  _\__, /  
                      /____/                           /____/   
` + "\n"

// StartService begins the server and allows the player to connect to it
// The player connects and once that player connects game data is sent to the server
func StartService(){
	var config ServerConfig
	// this will open our config file name is config.toml
	f, err := os.Open("config.toml")
	if err != nil {
		panic(err)
	}
	if err := toml.NewDecoder(f).Decode(&config); err != nil {
		panic(err)
	}
	fmt.Println(ascii)
	listener, err := minecraft.Listen("raknet", config.Address)
	listener.ServerName = config.ServerName
	listener.MaximumPlayers = config.MaxPlayers
	logger.LogInfo("Starting server...")
	logger.LogInfo("Server has started!")
	if err != nil {
		panic(err)
	}

	for {
		c, err := listener.Accept()
		if err != nil {
			return
		}
		conn := c.(*minecraft.Conn)

		go func() {
			defer conn.Close()
			// Sends game data
			data := minecraft.GameData{
				WorldName:       config.WorldName,
				EntityUniqueID:  0,
				EntityRuntimeID: 0,
				PlayerGameMode:  0,
				PlayerPosition:  mgl32.Vec3{},
				Pitch:           0,
				Yaw:             0,
				Dimension:       0,
				WorldSpawn:      protocol.BlockPos{},
				GameRules:       nil,
				Time:            0,
				Blocks:          nil,
				Items:           nil,
			}
			if err := conn.StartGame(data); err != nil {
				return
			}
		}()
	}
}
