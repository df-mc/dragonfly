package chat

import (
	"fmt"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"strings"
)

// Subscriber represents an entity that may subscribe to a Chat. In order to do so, the Subscriber must
// implement methods to send messages to it.
type Subscriber interface {
	// Message sends a formatted message to the subscriber. The message is formatted as it would when using
	// fmt.Println.
	Message(a ...any)
}

// StdoutSubscriber is an implementation of Subscriber that forwards messages sent to the chat to the stdout.
type StdoutSubscriber struct{}

// Message ...
func (c StdoutSubscriber) Message(a ...any) {
	s := make([]string, len(a))
	for i, b := range a {
		s[i] = fmt.Sprint(b)
	}
	t := text.ANSI(strings.Join(s, " "))
	if !strings.HasSuffix(t, "\n") {
		fmt.Println(t)
		return
	}
	fmt.Print(t)
}
