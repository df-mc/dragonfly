package dragonfly

// statusProvider handles the way the server shows up in the server list. The online players and maximum
// players are not changeable from outside of the server, but the server name may be changed at any time.
type statusProvider struct {
	s *Server
}

// ServerStatus provides the server status to the minecraft.Listener.
func (s statusProvider) ServerStatus() (name string, onlinePlayers, maxPlayers int) {
	s.s.nameMu.Lock()
	defer s.s.nameMu.Unlock()

	return s.s.name, s.s.PlayerCount(), s.s.MaxPlayerCount()
}
