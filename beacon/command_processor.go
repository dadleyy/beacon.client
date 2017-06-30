package beacon

import "sync"
import "bytes"
import "crypto"
import "crypto/rand"
import "encoding/hex"
import "github.com/hink/go-blink1"
import "github.com/golang/protobuf/proto"

import "github.com/dadleyy/beacon.client/beacon/defs"
import "github.com/dadleyy/beacon.client/beacon/logging"
import "github.com/dadleyy/beacon.client/beacon/interchange"

// NewCommandProcessor builds a new command processor w/ a default logger
func NewCommandProcessor(device Commandable, key crypto.Decrypter, stream <-chan *bytes.Buffer) *CommandProcessor {
	logger := logging.New(defs.CommandProcessorLoggerPrefix, logging.Magenta)
	return &CommandProcessor{logger, key, device, stream}
}

// CommandProcessor defines the main background processor that receives device messages and sends them to the device
type CommandProcessor struct {
	*logging.Logger
	crypto.Decrypter

	device        Commandable
	commandStream <-chan *bytes.Buffer
}

// Start initiates the reading of the command stream
func (processor *CommandProcessor) Start(wg *sync.WaitGroup) {
	defer wg.Done()
	processor.Infof("command processor starting")

	for buffer := range processor.commandStream {
		message := &interchange.DeviceMessage{}

		if e := proto.UnmarshalMerge(buffer.Bytes(), message); e != nil {
			processor.Warnf("unable to unmarshal protobuf message: %s", e.Error())
			continue
		}

		auth := message.GetAuthentication()

		if auth == nil {
			processor.Warnf("received message missing authentication information, continuing")
			continue
		}

		digestBytes, e := hex.DecodeString(auth.MessageDigest)

		if e != nil {
			processor.Warnf("invalid hex message digest: %s, received:\n%s\n", e.Error(), auth.MessageDigest)
			continue
		}

		if _, e := processor.Decrypt(rand.Reader, digestBytes, nil); e != nil {
			processor.Warnf("unable to decode message digest from processor: %s, received:\n%s\n", e.Error(), auth.MessageDigest)
			continue
		}

		processor.Debugf("received message digest: %s", auth.MessageDigest[0:7])
		switch message.Type {
		case interchange.DeviceMessageType_WELCOME:
			welcome := &interchange.WelcomeMessage{}

			if e := proto.Unmarshal(message.GetPayload(), welcome); e != nil {
				processor.Warnf("unable to unmarshal welcome payload: %s", e.Error())
				continue
			}

			processor.Debugf("received welcome message: %v", welcome.DeviceID)
		case interchange.DeviceMessageType_CONTROL:
			control := &interchange.ControlMessage{}

			if e := proto.Unmarshal(message.GetPayload(), control); e != nil {
				processor.Debugf("unable to unmarshal control payload: %s", e.Error())
				continue
			}

			if control == nil || len(control.Frames) == 0 {
				processor.Debugf("skipping control message, no valid frames")
				continue
			}

			go processor.execute(control)
		default:
			processor.device.SetState(blink1.State{})
		}
	}
}

func (processor *CommandProcessor) execute(control *interchange.ControlMessage) {
	processor.Debugf("received control message w/ %d frames", len(control.Frames))

	for _, frame := range control.Frames {
		state := blink1.State{
			Blue:  uint8(frame.Blue),
			Red:   uint8(frame.Red),
			Green: uint8(frame.Green),
		}

		processor.device.SetState(state)
	}
}
