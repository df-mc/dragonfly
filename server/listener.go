package server

import (
	"cmp"
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
	"path/filepath"
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

// importPrivateKey reads a PEM file containing a P-384 [ecdsa.PrivateKey] and
// returns it for use by the NetherNet listener.
func importPrivateKey(path string) (*ecdsa.PrivateKey, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err // already wrapped in os.PathError
	}
	block, _ := pem.Decode(b)
	if block == nil {
		return nil, errors.New("invalid PEM block")
	}
	var key *ecdsa.PrivateKey
	switch block.Type {
	case "EC PRIVATE KEY":
		key, err = x509.ParseECPrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("parse private key: %w", err)
		}
	case "PRIVATE KEY":
		parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("parse private key: %w", err)
		}
		var ok bool
		key, ok = parsed.(*ecdsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("must be *ecdsa.PrivateKey: %T", parsed)
		}
	default:
		return nil, fmt.Errorf("invalid block type: %s", block.Type)
	}
	if key.Curve != elliptic.P384() {
		return nil, fmt.Errorf("private key must use P-384, got %s", key.Curve.Params().Name)
	}
	return key, nil
}

// exportPrivateKey writes a PEM file containing the [ecdsa.PrivateKey].
func exportPrivateKey(path string, key *ecdsa.PrivateKey) error {
	keyBytes, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return fmt.Errorf("encode: %w", err)
	}
	b := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: keyBytes,
	})
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("make parent directories: %w", err)
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	if _, err := f.Write(b); err != nil {
		_ = f.Close()
		_ = os.Remove(path)
		return fmt.Errorf("write: %w", err)
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(path)
		return fmt.Errorf("close: %w", err)
	}
	return nil
}

// netherNetKey returns the private key identifying the NetherNet listener, imported from path
// if it exists and generated otherwise. Generated keys are persisted to path so that the server
// identity survives restarts; an empty path yields an unsaved temporary key.
func netherNetKey(path string, log *slog.Logger) (*ecdsa.PrivateKey, error) {
	if path == "" {
		key, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("generate key: %w", err)
		}
		log.Warn("Using a temporary private key for the NetherNet listener. Players connecting over plain HTTP may see the TOFU (Trust On First Use) prompt every time the server restarts.")
		return key, nil
	}
	key, err := importPrivateKey(path)
	if os.IsNotExist(err) {
		key, err = ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("generate key: %w", err)
		}
		// Save the generated key so that players are not prompted to trust a
		// new server identity after every restart.
		if err := exportPrivateKey(path, key); err != nil {
			return nil, fmt.Errorf("export private key: %w", err)
		}
		log.Info("Generated a private key for NetherNet listener.", "path", path)
	} else if err != nil {
		return nil, fmt.Errorf("import key file: %w", err)
	}
	return key, nil
}

// netherNetListenerFunc may be used to return a *minecraft.Listener accepting NetherNet
// connections, negotiated over a plaintext HTTP signaling endpoint served on the configured
// address. It is used when UserConfig.Config() is called with a NetherNet transport enabled.
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
	key, err := netherNetKey(nnConf.KeyFile, lcfg.Log)
	if err != nil {
		return nil, err
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
		Handler:           logHTTPRequests(log, handler),
		ReadHeaderTimeout: cmp.Or(nnConf.ReadHeaderTimeout, 5*time.Second),
		ReadTimeout:       cmp.Or(nnConf.ReadTimeout, 10*time.Second),
		IdleTimeout:       cmp.Or(nnConf.IdleTimeout, 30*time.Second),
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

// logHTTPRequests logs signaling requests at debug level before passing them to next. The
// endpoint is publicly reachable, so anything louder would let pings and scanners spam the log.
func logHTTPRequests(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debug("NetherNet HTTP request.", "method", r.Method, "path", r.URL.Path, "raddr", r.RemoteAddr)
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
		Allow:                  conf.Allower.Allow,
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
	close func() error // stops the sidecar HTTP signaling server, if any
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
