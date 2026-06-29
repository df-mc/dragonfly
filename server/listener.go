package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/df-mc/dragonfly/server/session"
	"github.com/df-mc/go-nethernet"
	"github.com/df-mc/go-nethernet/endpoint"
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
	cfg := listenerConfig(conf)
	l, err := cfg.Listen("raknet", uc.Network.Address)
	if err != nil {
		return nil, fmt.Errorf("create minecraft listener: %w", err)
	}
	conf.Log.Info("Listener running.", "addr", l.Addr())
	return listener{Listener: l}, nil
}

func (uc UserConfig) netherNetListenerFunc(conf Config) (Listener, error) {
	nnConf := uc.Network.NetherNet
	if (nnConf.CertificateFile == "") != (nnConf.KeyFile == "") {
		return nil, errors.New("create NetherNet listener: certificate and key files must both be set")
	}
	address := nnConf.Address
	if address == "" {
		address = uc.Network.Address
	}
	tcp, err := net.Listen("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("listen NetherNet HTTP: %w", err)
	}
	if nnConf.CertificateFile != "" {
		if _, err := tls.LoadX509KeyPair(nnConf.CertificateFile, nnConf.KeyFile); err != nil {
			_ = tcp.Close()
			return nil, fmt.Errorf("read NetherNet certificate key pair: %w", err)
		}
	}
	log := conf.Log.With("net origin", "nethernet-http")
	handler := endpoint.HandlerConfig{Logger: log}.New()
	cfg := listenerConfig(conf)
	l, err := cfg.ListenNetwork(minecraft.NetherNet{
		Signaling: handler,
		ListenConfig: nethernet.ListenConfig{
			Log:               conf.Log.With("net origin", "nethernet"),
			AllowAnonymous:    conf.AuthDisabled,
			DisableTrickleICE: true,
		},
	}, handler.NetworkID())
	if err != nil {
		_ = tcp.Close()
		return nil, fmt.Errorf("create NetherNet listener: %w", err)
	}

	httpServer := &http.Server{
		Addr:              address,
		Handler:           logHTTPRequests(log, handler),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
	go func() {
		var err error
		if nnConf.CertificateFile != "" {
			err = httpServer.ServeTLS(tcp, nnConf.CertificateFile, nnConf.KeyFile)
		} else {
			err = httpServer.Serve(tcp)
		}
		if err != nil && !errors.Is(err, http.ErrServerClosed) && !errors.Is(err, net.ErrClosed) {
			conf.Log.Error("NetherNet HTTP listener closed unexpectedly: " + err.Error())
		}
	}()
	conf.Log.Info("NetherNet listener running.", "addr", tcp.Addr(), "https", nnConf.CertificateFile != "")
	return listener{Listener: l, close: httpServer.Close}, nil
}

func logHTTPRequests(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info("NetherNet HTTP request.", "method", r.Method, "path", r.URL.Path, "raddr", r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

func listenerConfig(conf Config) minecraft.ListenConfig {
	cfg := minecraft.ListenConfig{
		MaximumPlayers:         conf.MaxPlayers,
		StatusProvider:         conf.StatusProvider,
		AuthenticationDisabled: conf.AuthDisabled,
		ResourcePacks:          conf.Resources,
		TexturePacksRequired:   conf.ResourcesRequired,
		Compression:            conf.Compression,
	}
	if conf.Log.Enabled(context.Background(), slog.LevelDebug) {
		cfg.ErrorLog = conf.Log.With("net origin", "gophertunnel")
	}
	return cfg
}

// listener is a Listener implementation that wraps around a minecraft.Listener so that it can be listened on by
// Server.
type listener struct {
	*minecraft.Listener
	close func() error
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

// Close closes the Minecraft listener and any sidecar listener it depends on.
func (l listener) Close() error {
	err := l.Listener.Close()
	if l.close != nil {
		if closeErr := l.close(); err == nil {
			err = closeErr
		}
	}
	return err
}
