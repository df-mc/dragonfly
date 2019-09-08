package chat

// Subscriber represents an entity that may subscribe to a Chat. In order to do so, the Subscriber must
// implement methods to send messages to it.
type Subscriber interface {
	// Message sends a formatted message to the subscriber. The message is formatted as it would when using
	// fmt.Println.
	Message(a ...interface{})
}
