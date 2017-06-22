package beacon

import "github.com/hink/go-blink1"

// Commandable defines the interface used by the command processor
type Commandable interface {
	Close()
	SetState(blink1.State) error
}
