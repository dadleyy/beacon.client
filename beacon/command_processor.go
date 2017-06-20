package beacon

import "log"
import "sync"
import "bytes"
import "github.com/hink/go-blink1"

type messageHeader struct {
	DeviceId string `json:"device_id"`
}

type CommandProcessor struct {
	*log.Logger
	Device        Commandable
	CommandStream <-chan *bytes.Buffer
}

func (processor *CommandProcessor) Start(wg *sync.WaitGroup) {
	defer wg.Done()

	for buffer := range processor.CommandStream {
		processor.Printf("received: %s", buffer)
		processor.Device.SetState(blink1.State{})
	}
}
