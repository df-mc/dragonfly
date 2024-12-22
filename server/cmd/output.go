package cmd

import (
	"errors"
	"fmt"
	"github.com/df-mc/dragonfly/server/player/chat"
)

// Output holds the output of a command execution. It holds success messages
// and error messages, which the source of a command execution gets sent.
type Output struct {
	errors   []error
	messages []fmt.Stringer
}

// Errorf formats an error message and adds it to the command output.
func (o *Output) Errorf(format string, a ...any) {
	o.errors = append(o.errors, fmt.Errorf(format, a...))
}

// Error formats an error message and adds it to the command output.
func (o *Output) Error(a ...any) {
	o.errors = append(o.errors, errors.New(fmt.Sprint(a...)))
}

// Errort adds a translation as an error message and parameterises it using the
// arguments passed. Errort panics if the number of arguments is incorrect.
func (o *Output) Errort(t chat.Translation, a ...any) {
	o.errors = append(o.errors, t.F(a...))
}

// Printf formats a (success) message and adds it to the command output.
func (o *Output) Printf(format string, a ...any) {
	o.messages = append(o.messages, stringer(fmt.Sprintf(format, a...)))
}

// Print formats a (success) message and adds it to the command output.
func (o *Output) Print(a ...any) {
	o.messages = append(o.messages, stringer(fmt.Sprint(a...)))
}

// Printt adds a translation as a (success) message and parameterises it using
// the arguments passed. Printt panics if the number of arguments is incorrect.
func (o *Output) Printt(t chat.Translation, a ...any) {
	o.messages = append(o.messages, t.F(a...))
}

// Errors returns a list of all errors added to the command output. Usually
// only one error message is set: After one error message, execution of a
// command typically terminates.
func (o *Output) Errors() []error {
	return o.errors
}

// ErrorCount returns the count of errors that the command output has.
func (o *Output) ErrorCount() int {
	return len(o.errors)
}

// Messages returns a list of all messages added to the command output. The
// amount of messages present depends on the command called.
func (o *Output) Messages() []fmt.Stringer {
	return o.messages
}

// MessageCount returns the count of (success) messages that the command output
// has.
func (o *Output) MessageCount() int {
	return len(o.messages)
}

type stringer string

func (s stringer) String() string { return string(s) }
