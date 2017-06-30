package beacon

import "sync"
import "time"

import "github.com/dadleyy/beacon.client/beacon/defs"
import "github.com/dadleyy/beacon.client/beacon/logging"

// NewHeartbeatProcessor creates a new processor for heartbeats
func NewHeartbeatProcessor(pinger Pingable, delay time.Duration, retries uint) *HeartbeatProcessor {
	logger := logging.New(defs.HeartbeatProcessorLoggerPrefix, logging.Cyan)
	return &HeartbeatProcessor{logger, delay, pinger, retries}
}

// HeartbeatProcessor is responsible for keeping the websocket connection alive
type HeartbeatProcessor struct {
	*logging.Logger
	delay      time.Duration
	pinger     Pingable
	maxRetries uint
}

// Start launches the hearbeat sequence
func (processor *HeartbeatProcessor) Start(wg *sync.WaitGroup) {
	defer wg.Done()
	ticker, retries := time.NewTicker(processor.delay), 0
	processor.Infof("heartbeat processor starting")

	for _ = range ticker.C {
		e := processor.pinger.Ping([]byte("ping"))

		if e != nil && retries < 100 {
			retries++
			processor.Errorf("error pinging, retrying #%d in %f seconds (%s)", retries, processor.delay.Seconds(), e.Error())
			time.Sleep(processor.delay)
			continue
		}

		if e != nil {
			processor.Errorf("unable to open up writer: %s", e.Error())
			ticker.Stop()
			break
		}

		processor.Debugf("successfully pinged api host")
	}

}
