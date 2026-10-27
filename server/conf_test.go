package server

import (
	"fmt"
	"log/slog"
	"testing"

	"github.com/df-mc/dragonfly/server/world"
	"github.com/pelletier/go-toml"
)

func TestBlockNetworkIDConfig(t *testing.T) {
	uc := DefaultConfig()
	if !uc.Network.BlockNetworkIDHashes {
		t.Fatal("default config must use block network ID hashes")
	}
	uc.World.SaveData, uc.Players.SaveData = false, false
	uc.Resources.Folder = t.TempDir()
	w := world.Config{}.New()
	defer w.Close()
	for _, hashes := range []bool{false, true} {
		t.Run(fmt.Sprintf("hashes=%t", hashes), func(t *testing.T) {
			if err := toml.Unmarshal([]byte(fmt.Sprintf("[Network]\nBlockNetworkIDHashes = %t\n", hashes)), &uc); err != nil {
				t.Fatal(err)
			}
			conf, err := uc.Config(slog.New(slog.DiscardHandler))
			if err != nil {
				t.Fatal(err)
			}
			if conf.DisableBlockNetworkIDHashes != !hashes {
				t.Fatalf("DisableBlockNetworkIDHashes = %t, want %t", conf.DisableBlockNetworkIDHashes, !hashes)
			}
			srv := &Server{conf: conf, world: w}
			if got := srv.defaultGameData().UseBlockNetworkIDHashes; got != hashes {
				t.Fatalf("StartGame hash mode = %t, want %t", got, hashes)
			}
		})
	}
	srv := &Server{world: w}
	if !srv.defaultGameData().UseBlockNetworkIDHashes {
		t.Fatal("zero-value server config must use block network ID hashes")
	}
}
