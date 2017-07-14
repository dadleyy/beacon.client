package beacon

import "fmt"
import "sync"
import "bytes"
import "net/url"
import "net/http"
import "crypto/rsa"
import "crypto/rand"
import "encoding/hex"
import "crypto/sha256"
import "github.com/hink/go-blink1"
import "github.com/golang/protobuf/proto"

import "github.com/dadleyy/beacon.client/beacon/defs"
import "github.com/dadleyy/beacon.client/beacon/logging"
import "github.com/dadleyy/beacon.client/beacon/interchange"

// Feedback defines the structure of messages sent from the command processor to the feedback processor.
type Feedback struct {
	Registration *RegistrationInfo
	Error        error
	State        blink1.State
}

type publishRequest struct {
	payloadData  []byte
	payloadType  interchange.FeedbackMessageType
	registration *RegistrationInfo
}

// NewFeedbackProcessor constructs a feedback processor
func NewFeedbackProcessor(stream <-chan *Feedback, apiHome url.URL) Processor {
	logger := logging.New(defs.FeedbackProcessorLoggerPrefix, logging.Blue)
	return &FeedbackProcessor{logger, stream, apiHome}
}

// FeedbackProcessor communicates back to the api the current state of the device
type FeedbackProcessor struct {
	logging.Logger

	stream  <-chan *Feedback
	apiHome url.URL
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

func (processor *FeedbackProcessor) publishReport(message *Feedback) {
	payload, e := proto.Marshal(&interchange.ReportMessage{
		Red:   uint32(message.State.Red),
		Green: uint32(message.State.Green),
		Blue:  uint32(message.State.Blue),
	})

	if e != nil {
		processor.Errorf("unable to marshal report: %s", e.Error())
		return
	}

	req := &publishRequest{
		payloadData:  payload,
		payloadType:  interchange.FeedbackMessageType_REPORT,
		registration: message.Registration,
	}

	if e := processor.publishPayload(req); e != nil {
		processor.Errorf("unable to publish report: %s", e.Error())
	}

	processor.Infof("successfully published report")
}

func (processor *FeedbackProcessor) publishError(message *Feedback) {
}

func (processor *FeedbackProcessor) publishPayload(request *publishRequest) error {
	s := sha256.New()

	if _, e := s.Write(request.payloadData); e != nil {
		return e
	}

	digest := bytes.NewBuffer([]byte{})
	signed, e := processor.sign(request.registration.serverKey, s.Sum(nil))

	if e != nil {
		return e
	}

	if _, e = digest.Write(signed); e != nil {
		return e
	}

	// Encode the digest to hex.
	digestString := hex.EncodeToString(digest.Bytes())

	payload, e := proto.Marshal(&interchange.FeedbackMessage{
		Type: request.payloadType,
		Authentication: &interchange.DeviceMessageAuthentication{
			MessageDigest: digestString,
			DeviceID:      request.registration.deviceID,
		},
		Payload: request.payloadData,
	})

	if e != nil {
		return e
	}

	buf := bytes.NewBuffer(payload)
	response, err := http.Post(processor.apiEndpoint(), defs.APIFeedbackContentTypeHeader, buf)

	if err != nil {
		return err
	}

	if response.StatusCode != 200 {
		return fmt.Errorf("invalid response from server: %d", response.StatusCode)
	}

	return nil
}

func (processor *FeedbackProcessor) apiEndpoint() string {
	u, _ := url.Parse(processor.apiHome.String())
	u.Path = defs.APIFeedbackEndpoint
	return u.String()
}

func (processor *FeedbackProcessor) sign(key *rsa.PublicKey, data []byte) ([]byte, error) {
	return rsa.EncryptOAEP(sha256.New(), rand.Reader, key, data, []byte(defs.APIReportMessageLabel))
}
