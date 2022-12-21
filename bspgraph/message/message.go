package message

// Message is implemented by types that can be processed by a Queue
type Message interface {
	Type() string
}

// Queue is implemented by types that can serve as message queues
type Queue interface {
	Close() error
	Enqueue(msg Message) error
	PendingMessages() bool
	DiscardMessages() error
	Messages() Iterator
}

// Iterator provides an API for iterating a list of messages
type Iterator interface {
	Next() bool
	Message() Message
	Error() error
}

// QueueFactory is a function that can create new Queue instances
type QueueFactory func() Queue
