package beacon

import "log"
import "sync"
import "bytes"
import "github.com/hink/go-blink1"
import "github.com/golang/protobuf/proto"
import "github.com/dadleyy/beacon.api/beacon/interchange"

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
		message := interchange.DeviceMessage{}

		if e := proto.Unmarshal(buffer.Bytes(), &message); e != nil {
			processor.Printf("receved strange message: %s", e.Error())
			continue
		}

		processor.Printf("received: %s", message.RequestPath)
		processor.Device.SetState(blink1.State{Red: 255})
	}
}
