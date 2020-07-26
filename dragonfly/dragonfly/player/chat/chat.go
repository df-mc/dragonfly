package chat

import (
	"fmt"
	"sync"
)

// Global represents a global chat. Players will write in this chat by default when they send any message in
// the chat.
var Global = New()

// Chat represents the in-game chat. Messages may be written to it to send a message to all subscribers. The
// zero value of Chat is a chat ready to use.
// Methods on Chat may be called from multiple goroutines concurrently.
type Chat struct {
	m           sync.Mutex
	subscribers []Subscriber
}

// New returns a new chat. A zero value is, however, usually sufficient for use.
func New() *Chat {
	return &Chat{}
}

// Println prints a formatted message to the chat. The message is formatted according to the fmt.Sprintln
// formatting rules.
func (chat *Chat) Println(a ...interface{}) {
	msg := fmt.Sprintln(a...)

	chat.m.Lock()
	defer chat.m.Unlock()

	for _, subscriber := range chat.subscribers {
		subscriber.Message(msg)
	}
}

// Printf prints a formatted message using a specific format to the chat. The message is formatted according
// to the fmt.Sprintf formatting rules.
func (chat *Chat) Printf(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	_, _ = chat.WriteString(msg)
}

// Write writes the byte slice p as a string to the chat. It is equivalent to calling
// Chat.WriteString(string(p)).
func (chat *Chat) Write(p []byte) (n int, err error) {
	return chat.WriteString(string(p))
}

// Subscribe adds a subscriber to the chat, sending it every message written to the chat. In order to remove
// it again, use Chat.Unsubscribe().
func (chat *Chat) Subscribe(s Subscriber) {
	chat.m.Lock()
	defer chat.m.Unlock()
	chat.subscribers = append(chat.subscribers, s)
}

// Unsubscribe removes a subscriber from the chat, so that messages written to the chat will no longer be
// sent to it.
func (chat *Chat) Unsubscribe(s Subscriber) {
	chat.m.Lock()
	defer chat.m.Unlock()

	if len(chat.subscribers) == 0 {
		chat.subscribers = nil
		return
	}
	n := make([]Subscriber, 0, len(chat.subscribers)-1)
	for _, subscriber := range chat.subscribers {
		if subscriber != s {
			n = append(n, subscriber)
		}
	}
	chat.subscribers = n
}

// WriteString writes a string s to the chat.
func (chat *Chat) WriteString(s string) (n int, err error) {
	chat.m.Lock()
	defer chat.m.Unlock()

	for _, subscriber := range chat.subscribers {
		subscriber.Message(s)
	}
	return len(s), nil
}

// Close closes the chat, removing all subscribers from it.
func (chat *Chat) Close() error {
	chat.m.Lock()
	chat.subscribers = nil
	chat.m.Unlock()
	return nil
}
