package beacon

import "os"
import "log"
import "sync"
import "bytes"
import "crypto"
import "crypto/rand"
import "github.com/hink/go-blink1"
import "github.com/golang/protobuf/proto"

import "github.com/dadleyy/beacon.client/beacon/defs"
import "github.com/dadleyy/beacon.client/beacon/interchange"

// NewCommandProcessor builds a new command processor w/ a default logger
func NewCommandProcessor(device Commandable, key crypto.Decrypter, stream <-chan *bytes.Buffer) *CommandProcessor {
	logger := log.New(os.Stdout, defs.CommandProcessorLoggerPrefix, defs.DefaultLogFlags)
	return &CommandProcessor{logger, key, device, stream}
}

// CommandProcessor defines the main background processor that receives device messages and sends them to the device
type CommandProcessor struct {
	*log.Logger
	crypto.Decrypter

	device        Commandable
	commandStream <-chan *bytes.Buffer
}

// Start initiates the reading of the command stream
func (processor *CommandProcessor) Start(wg *sync.WaitGroup) {
	defer wg.Done()
	processor.Printf("command processor starting")

	for buffer := range processor.commandStream {
		message := &interchange.DeviceMessage{}
		decodedMessage, e := processor.Decrypt(rand.Reader, buffer.Bytes(), nil)

		if e != nil {
			processor.Printf("[WARN] unable to decode message from processor: %s", e.Error())
			continue
		}

		if e := proto.UnmarshalMerge(decodedMessage, message); e != nil {
			processor.Printf("[WARN] unable to unmarshal protobuf message: %s", e.Error())
			continue
		}

		switch message.Type {
		case interchange.DeviceMessageType_WELCOME:
			welcome := &interchange.WelcomeMessage{}

			if e := proto.Unmarshal(message.GetPayload(), welcome); e != nil {
				processor.Printf("unable to unmarshal welcome payload: %s", e.Error())
				continue
			}

			processor.Printf("received welcome message: %v", welcome.DeviceID)
		case interchange.DeviceMessageType_CONTROL:
			control := &interchange.ControlMessage{}

			if e := proto.Unmarshal(message.GetPayload(), control); e != nil {
				processor.Printf("unable to unmarshal control payload: %s", e.Error())
				continue
			}

			if control == nil || len(control.Frames) == 0 {
				processor.Printf("skipping control message, no valid frames")
				continue
			}

			go processor.execute(control)
		default:
			processor.device.SetState(blink1.State{})
		}
	}
}

func (processor *CommandProcessor) execute(control *interchange.ControlMessage) {
	processor.Printf("received control message w/ %d frames", len(control.Frames))

	for _, frame := range control.Frames {
		state := blink1.State{
			Blue:  uint8(frame.Blue),
			Red:   uint8(frame.Red),
			Green: uint8(frame.Green),
		}

		processor.device.SetState(state)
	}
}
