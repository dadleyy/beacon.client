package beacon

import "io"

// Subscriber defines an interface that is used to connect and receive data
type Subscriber interface {
	Connected() bool
	ReadInto(io.Writer) error
	Connect() error
	Close() error
	Preregister(string) error
	Ping([]byte) error
}

// Pingable defines an interface that defines a Ping method
type Pingable interface {
	Ping([]byte) error
}
