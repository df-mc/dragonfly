package server

import (
	"github.com/sandertv/gophertunnel/minecraft"
)

// statusProvider handles the way the server shows up in the server list. The online players and maximum
// players are not changeable from outside the server, but the server name may be changed at any time.
type statusProvider struct {
	s *Server
}

// ServerStatus returns the player count, max players and the server's name as a minecraft.ServerStatus.
func (s statusProvider) ServerStatus(playerCount, maxPlayers int) minecraft.ServerStatus {
	return minecraft.ServerStatus{
		ServerName:  s.s.name.Load(),
		PlayerCount: playerCount,
		MaxPlayers:  maxPlayers,
	}
}
