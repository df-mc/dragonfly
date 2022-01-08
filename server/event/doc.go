// Package event exposes a single exported `Context` type that may be used to influence the execution flow of events
// that occur on a server.
// Generally, the caller of `event.C()` calls `Context.Stop()` or `Context.Continue()` to call a function when an event
// is or is not cancelled respectively. Such events may be cancelled by passing the created `Context` to the end user,
// who is then able to cancel it by calling `Context.Cancel()`.
// Additionally, a `Context.After()` function is exported, which may be used to call code after the `Context.Stop()` and
// `Context.Continue()` functions are called. Code performing the event should be run in these functions, so that the
// `Context.After()` function is able to run code of the end user immediately after the event itself happens.
package event
