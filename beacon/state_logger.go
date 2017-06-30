package beacon

import "github.com/hink/go-blink1"
import "github.com/dadleyy/beacon.client/beacon/logging"

// StateLogger implements the Commandable interface for debugging purposes
type StateLogger struct {
	logging.Logger
}

// Close no-op
func (logger *StateLogger) Close() {}

// SetState logs out the state received by the "device"
func (logger *StateLogger) SetState(state blink1.State) error {
	logger.Debugf("received rgb(%d,%d,%d)", state.Red, state.Green, state.Blue)
	return nil
}
