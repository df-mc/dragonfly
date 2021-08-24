package server

import (
	"github.com/df-mc/dragonfly/server/session"
	"github.com/sandertv/gophertunnel/minecraft"
	"io"
)

type Listener interface {
	Accept() (session.Conn, error)
	Disconnect(conn session.Conn, reason string) error
	io.Closer
}

type listener struct {
	*minecraft.Listener
}

func (l listener) Accept() (session.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	return conn.(session.Conn), err
}

func (l listener) Disconnect(conn session.Conn, reason string) error {
	return l.Listener.Disconnect(conn.(*minecraft.Conn), reason)
}
