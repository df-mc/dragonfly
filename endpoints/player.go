package endpoints

import (
	"net/http"
)

// playerCount (/player_count) returns the player count of the server when called.
func (s server) playerCount(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]interface{}{"count": s.s.PlayerCount()})
}

// maxPlayerCount (/max_player_count) returns the maximum player count of the server when called.
func (s server) maxPlayerCount(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]interface{}{"max_count": s.s.MaxPlayerCount()})
}

// players (/players) returns a list of data of all players currently playing on the server.
func (s server) players(w http.ResponseWriter, r *http.Request) {
	v := map[string][]map[string]interface{}{"players": {}}
	for _, player := range s.s.Players() {
		v["players"] = append(v["players"], map[string]interface{}{
			"uuid":     player.UUID().String(),
			"xuid":     player.XUID(),
			"username": player.Name(),
		})
	}
	writeJSON(w, v)
}
