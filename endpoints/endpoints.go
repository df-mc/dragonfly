package endpoints

import (
	"encoding/json"
	"github.com/dragonfly-tech/dragonfly/dragonfly"
	"net/http"
)

// Serve serves all dragonfly endpoints at the address passed. The endpoints return data which is taken from
// the dragonfly.Server passed, such as the player count.
func Serve(addr string, s *dragonfly.Server) error {
	server := server{s: s}

	http.HandleFunc("/player_count", server.playerCount)
	http.HandleFunc("/max_player_count", server.maxPlayerCount)
	http.HandleFunc("/players", server.players)

	return http.ListenAndServe(addr, nil)
}

// server holds the Dragonfly server. Endpoint implementations are implemented as receivers of this struct.
type server struct {
	s *dragonfly.Server
}

// writeJSON writes an object to the ResponseWriter as JSON data. It panics if for some reason the data could
// not be encoded.
func writeJSON(w http.ResponseWriter, data interface{}) {
	b, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	_, _ = w.Write(b)
}
