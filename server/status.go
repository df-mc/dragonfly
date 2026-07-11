package server

import (
	"sync/atomic"

	"github.com/sandertv/gophertunnel/minecraft"
)

// statusProvider handles the way the server shows up in the server list. The
// online players and maximum players are not changeable from outside the
// server, but the server name may be changed at any time.
type statusProvider struct {
	name string
}

// ServerStatus returns the player count, max players and the server's name as
// a minecraft.ServerStatus.
func (s statusProvider) ServerStatus(playerCount, maxPlayers int) minecraft.ServerStatus {
	return minecraft.ServerStatus{
		ServerName:  s.name,
		PlayerCount: playerCount,
		MaxPlayers:  maxPlayers,
	}
}

type sharedStatusProvider struct {
	provider    minecraft.ServerStatusProvider
	playerCount *atomic.Int64
}

func (s sharedStatusProvider) ServerStatus(_ int, maxPlayers int) minecraft.ServerStatus {
	return s.provider.ServerStatus(int(s.playerCount.Load()), maxPlayers)
}
