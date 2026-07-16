package server

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
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

// importPrivateKey reads an PEM file containing an [ecdsa.PrivateKey] and returns
// it for use by the NetherNet listener.
func importPrivateKey(path string) (*ecdsa.PrivateKey, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err // already wrapped in os.PathError
	}
	block, _ := pem.Decode(b)
	if block == nil {
		return nil, errors.New("invalid PEM block")
	}
	switch block.Type {
	case "EC PRIVATE KEY":
		return x509.ParseECPrivateKey(block.Bytes)
	case "PRIVATE KEY":
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("parse private key: %w", err)
		}
		k, ok := key.(*ecdsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("must be *ecdsa.PrivateKey: %T", key)
		}
		return k, nil
	default:
		return nil, fmt.Errorf("invalid block type: %s", block.Type)
	}
}

// exportPrivateKey writes an PEM file containing the [ecdsa.PrivateKey].
func exportPrivateKey(path string, key *ecdsa.PrivateKey) error {
	keyBytes, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return fmt.Errorf("encode: %w", err)
	}
	return os.WriteFile(path, pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: keyBytes,
	}), 0644)
}

func (uc UserConfig) netherNetListenerFunc(conf Config) (Listener, error) {
	nnConf := uc.Network.NetherNet
	address := nnConf.Address
	if address == "" {
		address = uc.Network.Address
	}
	if nnConf.Domain == "" {
		nnConf.Domain = "self"
	}

	lcfg := nethernet.ListenConfig{
		Log:            conf.Log.With("net origin", "nethernet"),
		AllowAnonymous: conf.AuthDisabled,
	}
	var key *ecdsa.PrivateKey
	if nnConf.KeyFile != "" {
		var err error
		key, err = importPrivateKey(nnConf.KeyFile)
		if os.IsNotExist(err) {
			key, err = ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
			if err != nil {
				return nil, fmt.Errorf("generate key: %w", err)
			}
			// If we generated a new key for the NetherNet listener, save it.
			// Otherwise, players may be prompted to trust the server identity
			// after every restart.
			if err := exportPrivateKey(nnConf.KeyFile, key); err != nil {
				return nil, fmt.Errorf("export private key: %w", err)
			}
			lcfg.Log.Info("Generated a private key for NetherNet listener.", "path", nnConf.KeyFile)
		} else if err != nil {
			return nil, fmt.Errorf("import key file: %w", err)
		}
	} else {
		var err error
		key, err = ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("generate key: %w", err)
		}
		lcfg.Log.Warn("Using a temporary private key for the NetherNet listener. Players connecting over plain HTTP may see the TOFU (Trust On First Use) prompt every time the server restarts.")
	}
	lcfg.IssueServerIdentity = func(ctx context.Context) (*nethernet.Identity, error) {
		return nethernet.GenerateServerIdentity(key, nnConf.Domain)
	}

	tcp, err := net.Listen("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("listen NetherNet HTTP: %w", err)
	}
	log := conf.Log.With("net origin", "nethernet-http")
	handler := endpoint.HandlerConfig{Logger: log}.New()
	cfg := listenerConfig(conf)
	l, err := cfg.ListenNetwork(minecraft.NetherNet{
		Signaling:    handler,
		ListenConfig: lcfg,
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
		err := httpServer.Serve(tcp)
		if err != nil && !errors.Is(err, http.ErrServerClosed) && !errors.Is(err, net.ErrClosed) {
			conf.Log.Error("NetherNet HTTP listener closed unexpectedly: " + err.Error())
		}
	}()
	conf.Log.Info("NetherNet listener running.", "addr", tcp.Addr())
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
