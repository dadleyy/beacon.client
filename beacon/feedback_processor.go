package beacon

import "sync"
import "bytes"
import "net/url"
import "crypto/rsa"
import "crypto/rand"
import "encoding/hex"
import "crypto/sha256"
import "github.com/hink/go-blink1"
import "github.com/golang/protobuf/proto"

import "github.com/dadleyy/beacon.client/beacon/defs"
import "github.com/dadleyy/beacon.client/beacon/logging"
import "github.com/dadleyy/beacon.client/beacon/interchange"

// FeedbackMessage defines the structure of messages sent from the command processor to the feedback processor.
type FeedbackMessage struct {
	Key   *rsa.PublicKey
	Error error
	State blink1.State
}

type publishRequest struct {
	payloadData []byte
	payloadType interchange.FeedbackMessageType
	key         *rsa.PublicKey
}

// NewFeedbackProcessor constructs a feedback processor
func NewFeedbackProcessor(stream <-chan *FeedbackMessage, apiHome url.URL) Processor {
	logger := logging.New(defs.FeedbackProcessorLoggerPrefix, logging.Blue)
	return &FeedbackProcessor{logger, stream}
}

// FeedbackProcessor communicates back to the api the current state of the device
type FeedbackProcessor struct {
	logging.Logger

	stream <-chan *FeedbackMessage
}

// Start should be used as the target of a goroutine - kicks of receiving on channel
func (processor *FeedbackProcessor) Start(wg *sync.WaitGroup) {
	defer wg.Done()
	processor.Infof("starting feedback processor")

	for message := range processor.stream {
		processor.Debugf("received message on feedback stream, publishing to server, %v", message)

		if message.Error != nil {
			processor.publishError(message)
			continue
		}

		processor.publishReport(message)
	}
}

func (processor *FeedbackProcessor) publishReport(message *FeedbackMessage) {
	payload, e := proto.Marshal(&interchange.ReportMessage{
		Red:   uint32(message.State.Red),
		Green: uint32(message.State.Green),
		Blue:  uint32(message.State.Blue),
	})

	if e != nil {
		processor.Errorf("unable to marshal report: %s", e.Error())
		return
	}

	processor.publishPayload(&publishRequest{
		payloadData: payload,
		payloadType: interchange.FeedbackMessageType_REPORT,
		key:         message.Key,
	})
}

func (processor *FeedbackProcessor) publishPayload(request *publishRequest) {
	s := sha256.New()

	if _, e := s.Write(request.payloadData); e != nil {
		processor.Errorf("unable to write report payload into sha digest: %s", e.Error())
		return
	}

	digest := bytes.NewBuffer([]byte{})
	signed, e := processor.sign(request.key, s.Sum(nil))

	if e != nil {
		processor.Errorf("unable to write report payload into sha digest: %s", e.Error())
		return
	}

	if _, e = digest.Write(signed); e != nil {
		processor.Errorf("unable to write report payload into sha digest: %s", e.Error())
		return
	}

	// Encode the digest to hex.
	digestString := hex.EncodeToString(digest.Bytes())

	message := interchange.FeedbackMessage{
		Type: request.payloadType,
		Authentication: &interchange.DeviceMessageAuthentication{
			MessageDigest: digestString,
		},
		Payload: request.payloadData,
	}

	processor.Debugf("sending digest string: %s (%#v)", digestString, message)
}

func (processor *FeedbackProcessor) sign(key *rsa.PublicKey, data []byte) ([]byte, error) {
	return rsa.EncryptOAEP(sha256.New(), rand.Reader, key, data, []byte(defs.APIReportMessageLabel))
}

func (processor *FeedbackProcessor) publishError(message *FeedbackMessage) {
}
