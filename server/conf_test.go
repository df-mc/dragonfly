package server

import (
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/df-mc/dragonfly/server/session"
)

// TestConfigNewSkipsFailedListener verifies that a listener that could not be created is not kept. A listener factory
// returns a nil Listener along with its error, and a nil listener that reaches the server is only found later, when a
// goroutine started for it accepts connections on it and dereferences it.
func TestConfigNewSkipsFailedListener(t *testing.T) {
	tests := []struct {
		name      string
		listeners []func(conf Config) (Listener, error)
		want      int
	}{
		{
			name:      "the only listener fails to bind",
			listeners: []func(conf Config) (Listener, error){failingListener},
			want:      0,
		},
		{
			name:      "one of two listeners fails to bind",
			listeners: []func(conf Config) (Listener, error){failingListener, nopListenerFunc},
			want:      1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := Config{
				Log:       slog.New(slog.NewTextHandler(io.Discard, nil)),
				Listeners: tt.listeners,
			}.New()
			if got := len(srv.listeners); got != tt.want {
				t.Fatalf("server holds %v listeners, want %v", got, tt.want)
			}
			for i, l := range srv.listeners {
				if l == nil {
					t.Errorf("listener %v is nil: accepting connections on it dereferences a nil interface", i)
				}
			}
		})
	}
}

// failingListener stands in for a listener that cannot bind, such as one whose port is already in use.
func failingListener(Config) (Listener, error) {
	return nil, errors.New("bind: address already in use")
}

func nopListenerFunc(Config) (Listener, error) {
	return nopListener{}, nil
}

type nopListener struct{}

func (nopListener) Accept() (session.Conn, error)         { return nil, io.EOF }
func (nopListener) Disconnect(session.Conn, string) error { return nil }
func (nopListener) Close() error                          { return nil }
