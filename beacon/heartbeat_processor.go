package beacon

import "os"
import "log"
import "sync"
import "time"

import "github.com/dadleyy/beacon.client/beacon/defs"

// NewHeartbeatProcessor creates a new processor for heartbeats
func NewHeartbeatProcessor(pinger Pingable, delay time.Duration) *HeartbeatProcessor {
	logger := log.New(os.Stdout, defs.HeartbeatProcessorLoggerPrefix, defs.DefaultLogFlags)
	return &HeartbeatProcessor{logger, delay, pinger}
}

// HeartbeatProcessor is responsible for keeping the websocket connection alive
type HeartbeatProcessor struct {
	*log.Logger
	delay  time.Duration
	pinger Pingable
}

// Start launches the hearbeat sequence
func (processor *HeartbeatProcessor) Start(wg *sync.WaitGroup) {
	defer wg.Done()
	ticker := time.NewTicker(processor.delay)
	processor.Printf("heartbeat processor starting")

	for _ = range ticker.C {
		e := processor.pinger.Ping([]byte("ping"))

		if e != nil {
			processor.Printf("unable to open up writer: %s", e.Error())
			ticker.Stop()
			break
		}

		processor.Printf("successfully pinged api host")
	}

}
