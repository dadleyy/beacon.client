package beacon

import "sync"

import "github.com/dadleyy/beacon.client/beacon/defs"
import "github.com/dadleyy/beacon.client/beacon/logging"

// NewFeedbackProcessor constructs a feedback processor
func NewFeedbackProcessor() Processor {
	logger := logging.New(defs.FeedbackProcessorLoggerPrefix, logging.Blue)
	return &FeedbackProcessor{logger}
}

// FeedbackProcessor communicates back to the api the current state of the device
type FeedbackProcessor struct {
	logging.Logger
}

// Start should be used as the target of a goroutine - kicks of receiving on channel
func (processor *FeedbackProcessor) Start(wg *sync.WaitGroup) {
	defer wg.Done()
	processor.Infof("starting feedback processor")
}
