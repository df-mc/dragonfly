package server

import (
	"github.com/sandertv/gophertunnel/minecraft"
)

// statusProvider handles the way the server shows up in the server list. The online players and maximum
// players are not changeable from outside of the server, but the server name may be changed at any time.
type statusProvider struct {
	s *Server
}

func (s statusProvider) ServerStatus(playerCount, maxPlayers int) minecraft.ServerStatus {
	return minecraft.ServerStatus{
		ServerName:  s.s.name.Load(),
		PlayerCount: playerCount,
		MaxPlayers:  maxPlayers,
	}
}
