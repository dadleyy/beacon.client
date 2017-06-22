package beacon

import "log"
import "github.com/hink/go-blink1"

// StateLogger implements the Commandable interface for debugging purposes
type StateLogger struct {
	*log.Logger
}

// Close no-op
func (logger *StateLogger) Close() {
}

// SetState logs out the state received by the "device"
func (logger *StateLogger) SetState(state blink1.State) error {
	logger.Printf("received rgb(%d,%d,%d)", state.Red, state.Green, state.Blue)
	return nil
}
