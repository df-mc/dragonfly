package internal

import "github.com/df-mc/dragonfly/server/world"

// Logger is a type that is used to log errors and info to throughout Dragonfly. Any logger implementation that
// implements Logger may be used by passing it to server.New.
type Logger interface {
	world.Logger
	Infof(format string, v ...any)
	Fatalf(format string, v ...any)
}
