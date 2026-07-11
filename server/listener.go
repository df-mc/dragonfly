package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"

	"github.com/df-mc/dragonfly/server/session"
	"github.com/sandertv/gophertunnel/minecraft"
)

// Listener is a source for connections that may be listened on by a Server using Server.listen. Proxies can use this to
// provide players from a different source.
type Listener interface {
	// Accept blocks until the next connection is established and returns it. An error is returned if the Listener was
	// closed using Close.
	Accept() (session.Conn, error)
	// Disconnect disconnects a connection from the Listener with a reason.
	Disconnect(conn session.Conn, reason string) error
	io.Closer
}

// listenerFunc may be used to return a *minecraft.Listener using a Config. It
// is the standard listener used when UserConfig.Config() is called.
func (uc UserConfig) listenerFunc(conf Config) (Listener, error) {
	cfg := minecraft.ListenConfig{
		MaximumPlayers:         conf.MaxPlayers,
		ListenerGroup:          new(minecraft.ListenerGroup),
		StatusProvider:         conf.StatusProvider,
		AuthenticationDisabled: conf.AuthDisabled,
		ResourcePacks:          conf.Resources,
		TexturePacksRequired:   conf.ResourcesRequired,
		Compression:            conf.Compression,
	}
	if conf.Log.Enabled(context.Background(), slog.LevelDebug) {
		cfg.ErrorLog = conf.Log.With("net origin", "gophertunnel")
	}
	primary, err := cfg.Listen("raknet", uc.Network.Address)
	if err != nil {
		return nil, fmt.Errorf("create minecraft listener: %w", err)
	}
	conf.Log.Info("Listener running.", "addr", primary.Addr())
	if uc.Network.AddressV6 == "" {
		return listener{primary}, nil
	}

	secondary, err := cfg.Listen("raknet", uc.Network.AddressV6)
	if err != nil {
		_ = primary.Close()
		return nil, fmt.Errorf("create IPv6 minecraft listener: %w", err)
	}
	conf.Log.Info("Listener running.", "addr", secondary.Addr())
	return newMultiListener(listener{primary}, listener{secondary}), nil
}

type acceptResult struct {
	conn session.Conn
	err  error
}

// multiListener combines multiple listeners into one so that the standard server listener can accept connections
// arriving over either IPv4 or IPv6.
type multiListener struct {
	listeners []Listener
	incoming  chan acceptResult
	done      chan struct{}
	closeOnce sync.Once
	closeErr  error
}

func newMultiListener(listeners ...Listener) *multiListener {
	l := &multiListener{listeners: listeners, incoming: make(chan acceptResult), done: make(chan struct{})}
	for _, child := range listeners {
		go func() {
			for {
				conn, err := child.Accept()
				select {
				case l.incoming <- acceptResult{conn: conn, err: err}:
				case <-l.done:
					return
				}
				if err != nil {
					_ = l.Close()
					return
				}
			}
		}()
	}
	return l
}

func (l *multiListener) Accept() (session.Conn, error) {
	select {
	case result := <-l.incoming:
		return result.conn, result.err
	case <-l.done:
		return nil, net.ErrClosed
	}
}

func (l *multiListener) Disconnect(conn session.Conn, reason string) error {
	return l.listeners[0].Disconnect(conn, reason)
}

func (l *multiListener) Close() error {
	l.closeOnce.Do(func() {
		close(l.done)
		err := make([]error, 0, len(l.listeners))
		for _, child := range l.listeners {
			err = append(err, child.Close())
		}
		l.closeErr = errors.Join(err...)
	})
	return l.closeErr
}

// listener is a Listener implementation that wraps around a minecraft.Listener so that it can be listened on by
// Server.
type listener struct {
	*minecraft.Listener
}

// Accept blocks until the next connection is established and returns it. An error is returned if the Listener was
// closed using Close.
func (l listener) Accept() (session.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	return conn.(session.Conn), err
}

// Disconnect disconnects a connection from the Listener with a reason.
func (l listener) Disconnect(conn session.Conn, reason string) error {
	return l.Listener.Disconnect(conn.(*minecraft.Conn), reason)
}
