// Package event exposes a single exported `Context` type that may be used to influence the execution flow of events
// that occur on a server.
// Generally, the caller of `event.C()` calls `Context.Cancelled()` to check if the `Context` was cancelled (using
// `Context.Cancel()`) by whatever code it was passed to.
// who is then able to cancel it by calling `Context.Cancel()`.
package event
