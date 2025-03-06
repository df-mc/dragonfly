package cmd

import (
	"errors"
	"fmt"
	"github.com/df-mc/dragonfly/server/player/chat"
	"golang.org/x/text/language"
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
	if len(a) == 1 {
		if err, ok := a[0].(error); ok {
			o.errors = append(o.errors, err)
			return
		}
	}
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

var MessageSyntax = chat.Translate(str("%commands.generic.syntax"), 3, `Syntax error: unexpected value: at "%v>>%v<<%v"`).Enc("<red>%v</red>")
var MessageUsage = chat.Translate(str("%commands.generic.usage"), 1, `Usage: %v`).Enc("<red>%v</red>")
var MessageUnknown = chat.Translate(str("%commands.generic.unknown"), 1, `Unknown command: "%v": Please check that the command exists and that you have permission to use it.`).Enc("<red>%v</red>")
var MessageNoTargets = chat.Translate(str("%commands.generic.noTargetMatch"), 0, `No targets matched selector`).Enc("<red>%v</red>")
var MessageNumberInvalid = chat.Translate(str("%commands.generic.num.invalid"), 1, `'%v' is not a valid number`).Enc("<red>> %v</red>")
var MessageBooleanInvalid = chat.Translate(str("%commands.generic.boolean.invalid"), 1, `'%v' is not true or false`).Enc("<red>> %v</red>")
var MessagePlayerNotFound = chat.Translate(str("%commands.generic.player.notFound"), 0, `That player cannot be found`).Enc("<red>> %v</red>")
var MessageParameterInvalid = chat.Translate(str("%commands.generic.parameter.invalid"), 1, `'%v' is not a valid parameter`).Enc("<red>> %v</red>")

type str string

// Resolve returns the translation identifier as a string.
func (s str) Resolve(language.Tag) string { return string(s) }
