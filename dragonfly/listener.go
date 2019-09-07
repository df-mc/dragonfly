package dragonfly

import (
	"fmt"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sirupsen/logrus"
	"log"
)

// Server implements a Dragonfly server. It runs the main server loop and handles the connections of players
// trying to join the server.
type Server struct {
	c   Config
	log *logrus.Logger

	listener *minecraft.Listener
}

// New returns a new server using the Config passed. If nil is passed, a default configuration is returned.
// (A call to dragonfly.DefaultConfig().)
// The Logger passed will be used to log errors and information to. If nil is passed, a default Logger is
// used by calling logrus.New().
func New(c *Config, log *logrus.Logger) *Server {
	if log == nil {
		log = logrus.New()
	}
	if c == nil {
		return &Server{c: DefaultConfig(), log: log}
	}
	return &Server{c: *c, log: log}
}

// Run runs the server and blocks until it is closed using a call to Close(). When called, the server will
// accept incoming connections.
func (server *Server) Run() error {
	server.log.Info("Starting server...")

	w := server.log.Writer()
	defer func() {
		_ = w.Close()
	}()

	server.listener = &minecraft.Listener{
		// We wrap a log.Logger around our Logrus logger so that it will print in the same format as the
		// normal Logrus logger would.
		ErrorLog:       log.New(w, "", 0),
		ServerName:     server.c.Server.Name,
		MaximumPlayers: server.c.Server.MaximumPlayers,
	}
	if err := server.listener.Listen("raknet", server.c.Network.Address); err != nil {
		return fmt.Errorf("listening on address failed: %v", err)
	}

	server.log.Infof("Server started on %v\n", server.c.Network.Address)

	for {
		c, err := server.listener.Accept()
		if err != nil {
			// Accept will only return an error if the Listener was closed, meaning trying to continue
			// listening is futile.
			return nil
		}
		go server.handleConn(c.(*minecraft.Conn))
	}
}

// handleConn handles an incoming connection accepted from the Listener.
func (server *Server) handleConn(conn *minecraft.Conn) {
	defer func() {
		_ = conn.Close()
	}()
	data := minecraft.GameData{WorldName: server.c.Server.WorldName}
	if err := conn.StartGame(data); err != nil {
		return
	}
	// TODO: Handle the connection.
}

// Close closes the server, making any call to Run cancel immediately.
func (server *Server) Close() error {
	return server.listener.Close()
}
