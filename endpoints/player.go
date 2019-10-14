package endpoints

import (
	"github.com/google/uuid"
	"net/http"
)

// playerCount (/player_count) returns the player count of the server when called.
func (s server) playerCount(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, map[string]interface{}{"count": s.s.PlayerCount()})
}

// maxPlayerCount (/max_player_count) returns the maximum player count of the server when called.
func (s server) maxPlayerCount(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, map[string]interface{}{"max_count": s.s.MaxPlayerCount()})
}

// players (/players) returns a list of data of all players currently playing on the server.
func (s server) players(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
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

// kick (/kick) kicks a player from the server. If the player is not currently online, no player is returned.
// The kick endpoint takes two form values:
//   "uuid": The UUID of the player to be kicked.
//   "message": The message that the player should be kicked with.
func (s server) kick(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	id, err := uuid.Parse(r.FormValue("uuid"))
	if err != nil {
		badRequest(w, "Malformed UUID passed", errMalformedUUID)
		return
	}
	player, found := s.s.Player(id)
	if !found {
		badRequest(w, "Player could not be found", errPlayerNotFound)
	}
	player.Disconnect(r.FormValue("message"))
}
