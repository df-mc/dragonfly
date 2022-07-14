package cmd

import "fmt"

// Output holds the output of a command execution. It holds success messages and error messages, which the
// source of a command execution gets sent.
type Output struct {
	errors   []error
	messages []string
}

// Errorf formats an error message and adds it to the command output.
func (o *Output) Errorf(format string, a ...any) {
	o.errors = append(o.errors, fmt.Errorf(format, a...))
}

// Error formats an error message and adds it to the command output.
func (o *Output) Error(a ...any) {
	o.errors = append(o.errors, fmt.Errorf(fmt.Sprint(a...)))
}

// Printf formats a (success) message and adds it to the command output.
func (o *Output) Printf(format string, a ...any) {
	o.messages = append(o.messages, fmt.Sprintf(format, a...))
}

// Print formats a (success) message and adds it to the command output.
func (o *Output) Print(a ...any) {
	o.messages = append(o.messages, fmt.Sprint(a...))
}

// Errors returns a list of all errors added to the command output. Usually only one error message is set:
// After one error message, execution of a command typically terminates.
func (o *Output) Errors() []error {
	return o.errors
}

// ErrorCount returns the count of errors that the command output has.
func (o *Output) ErrorCount() int {
	return len(o.errors)
}

// Messages returns a list of all messages added to the command output. The amount of messages present depends
// on the command called.
func (o *Output) Messages() []string {
	return o.messages
}

// MessageCount returns the count of (success) messages that the command output has.
func (o *Output) MessageCount() int {
	return len(o.messages)
}
