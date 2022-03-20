package internal

// Logger is a type that is used to log errors and info to throughout Dragonfly. Any logger implementation that
// implements Logger may be used by passing it to server.New.
type Logger interface {
	Debugf(format string, v ...any)
	Infof(format string, v ...any)
	Errorf(format string, v ...any)
	Fatalf(format string, v ...any)
}
