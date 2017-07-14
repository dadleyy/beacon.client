package beacon

import "fmt"
import "sync"
import "bytes"
import "crypto"
import "crypto/rsa"
import "crypto/rand"
import "crypto/x509"
import "encoding/hex"
import "github.com/google/uuid"
import "github.com/hink/go-blink1"
import "github.com/golang/protobuf/proto"

import "github.com/dadleyy/beacon.client/beacon/defs"
import "github.com/dadleyy/beacon.client/beacon/logging"
import "github.com/dadleyy/beacon.client/beacon/interchange"

// Decrypter is an alias for the crypto.Decrypter interface
type Decrypter crypto.Decrypter

// NewCommandProcessor builds a new command processor w/ a default logger.
func NewCommandProcessor(d Commandable, k Decrypter, c <-chan *bytes.Buffer, f chan<- *Feedback) Processor {
	l := logging.New(defs.CommandProcessorLoggerPrefix, logging.Magenta)
	return &CommandProcessor{l, k, d, c, f, nil, nil}
}

// CommandProcessor defines the main background processor that receives device messages and sends them to the device
type CommandProcessor struct {
	logging.Logger
	crypto.Decrypter

	device         Commandable
	commandStream  <-chan *bytes.Buffer
	feedbackStream chan<- *Feedback

	latestMessage *uuid.UUID
	registration  *RegistrationInfo
}

// Start initiates the reading of the command stream
func (processor *CommandProcessor) Start(wg *sync.WaitGroup) {
	defer wg.Done()
	processor.Infof("command processor starting")

	// Iterate over the command stream channel for as long as we're open.
	for buffer := range processor.commandStream {
		message := &interchange.DeviceMessage{}

		// Attempt to unmarshal the buffer we've received into our device message protocol buffer.
		if e := proto.UnmarshalMerge(buffer.Bytes(), message); e != nil {
			processor.Warnf("unable to unmarshal protobuf message: %s", e.Error())
			continue
		}

		// Validate our message based on our Decrypter interface + the authentication's digest.
		if e := processor.validateMessage(message); e != nil {
			processor.Warnf("unable to validate message: %s", e.Error())
			continue
		}

		processor.Debugf("received message digest: %s", message.Authentication.MessageDigest[0:7])

		// Decide which type of message this is.
		switch message.Type {
		case interchange.DeviceMessageType_WELCOME:
			// If we'reve receved a welcome message, we need to extract the server public key from the message contents.
			var e error
			processor.registration, e = processor.parseWelcomeMessage(message)

			if e != nil {
				processor.Warnf("incorrect shared secret key, not rsa format: %s", e.Error())
				continue
			}
		case interchange.DeviceMessageType_CONTROL:
			control := &interchange.ControlMessage{}

			// If we haven't received the server key, do nothing!
			if processor.registration == nil {
				processor.Warnf("have not received server key from welcome message, continuing")
				continue
			}

			// Attempt to unmarshal our message payload into our control message protocol buffer.
			if e := proto.Unmarshal(message.GetPayload(), control); e != nil {
				processor.Debugf("unable to unmarshal control payload: %s", e.Error())
				continue
			}

			// If we received a strange control message (empty or w/o any frames), skip it.
			if control == nil || len(control.Frames) == 0 {
				processor.Debugf("skipping control message, no valid frames")
				continue
			}

			controlID, e := uuid.NewUUID()

			if e != nil {
				processor.Warnf("unable to generate control id: %s", e.Error())
				continue
			}

			// Set the processor's latest message to allow execution interruption.
			processor.latestMessage = &controlID

			// Executre the control message in a goroutine, the previous attempt will terminate
			go processor.execute(control, &controlID)
		default:
			// If we do not understand the type of the message, turn the device off.
			processor.device.SetState(blink1.State{})
		}
	}
}

func (processor *CommandProcessor) validateMessage(message *interchange.DeviceMessage) error {
	// Access the authentication portion of our device message.
	auth := message.GetAuthentication()

	if auth == nil {
		processor.Warnf("received message missing authentication information, continuing")
		return fmt.Errorf("invalid-authentication")
	}

	digestBytes, e := hex.DecodeString(auth.MessageDigest)

	if e != nil {
		return e
	}

	if _, e := processor.Decrypt(rand.Reader, digestBytes, nil); e != nil {
		return e
	}

	return nil
}

func (processor *CommandProcessor) parseWelcomeMessage(message *interchange.DeviceMessage) (*RegistrationInfo, error) {
	welcome := &interchange.WelcomeMessage{}

	auth := message.GetAuthentication()

	if auth == nil {
		return nil, fmt.Errorf("invalid-message-auth")
	}

	if e := proto.Unmarshal(message.GetPayload(), welcome); e != nil {
		return nil, e
	}

	processor.Debugf("received welcome, deviceID[%s]", auth.DeviceID)
	block, e := hex.DecodeString(welcome.SharedSecret)

	if e != nil {
		return nil, e
	}

	pub, e := x509.ParsePKIXPublicKey(block)

	if e != nil {
		return nil, e
	}

	serverKey, ok := pub.(*rsa.PublicKey)

	if ok != true {
		return nil, fmt.Errorf("invalid-public-key")
	}

	return &RegistrationInfo{serverKey, auth.DeviceID}, nil
}

func (processor *CommandProcessor) execute(control *interchange.ControlMessage, id *uuid.UUID) {
	processor.Debugf("received control message w/ %d frames", len(control.Frames))

	for _, frame := range control.Frames {
		// If we have a latest message and it is not the same as the one we were given, skip everything.
		if processor.latestMessage != nil && processor.latestMessage.String() != id.String() {
			return
		}

		state := blink1.State{
			Blue:  uint8(frame.Blue),
			Red:   uint8(frame.Red),
			Green: uint8(frame.Green),
		}

		if e := processor.device.SetState(state); e != nil {
			processor.Errorf("unable to set device state, aborting control frames: %s", e.Error())
			processor.feedbackStream <- &Feedback{Registration: processor.registration, Error: e}
			return
		}

		processor.feedbackStream <- &Feedback{
			Registration: processor.registration,
			State:        state,
		}
	}

	// We're now done, let future executions know there is no currently executing control message.
	processor.latestMessage = nil
}
