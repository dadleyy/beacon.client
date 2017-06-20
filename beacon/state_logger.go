package beacon

import "log"
import "github.com/hink/go-blink1"

type StateLogger struct {
	*log.Logger
}

func (logger *StateLogger) Close() {
}

func (logger *StateLogger) SetState(state blink1.State) error {
	return nil
}
