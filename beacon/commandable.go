package beacon

import "github.com/hink/go-blink1"

type Commandable interface {
	Close()
	SetState(blink1.State) error
}
