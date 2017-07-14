package main

import "flag"
import "sync"
import "time"
import "bytes"
import "net/url"

import "github.com/hink/go-blink1"
import "github.com/dadleyy/beacon.client/beacon"
import "github.com/dadleyy/beacon.client/beacon/defs"
import "github.com/dadleyy/beacon.client/beacon/logging"
import "github.com/dadleyy/beacon.client/beacon/security"

func main() {
	options := struct {
		apiHome        string
		debugging      bool
		commandBuffer  int
		heartbeatDelay int
		privateKeyfile string
		deviceName     string
		maxRetries     int
		retryDelay     int
	}{}

	flag.StringVar(&options.apiHome, "api", "http://0.0.0.0:12345", "the hostname of the beacon.api server")
	flag.BoolVar(&options.debugging, "debug", false, "if true, the client will not attempt to open the blink device")
	flag.IntVar(&options.commandBuffer, "command-buffer", 2, "amount of allowed commands to buffer")
	flag.IntVar(&options.heartbeatDelay, "heartbeat-delay", 10, "amount of seconds between heartbeat pings")
	flag.IntVar(&options.retryDelay, "retry-delay", 5, "amount of seconds to wait before retrying")
	flag.IntVar(&options.maxRetries, "max-retries", 10, "amount of attempts the client will attempt to reconnect")
	flag.StringVar(&options.privateKeyfile, "privte-key", ".keys/private.pem", "the filename of the private key")
	flag.StringVar(&options.deviceName, "device-name", "", "if provided, this will attempt to pre-register with the api")
	flag.Parse()

	if len(options.apiHome) < 1 {
		flag.PrintDefaults()
		return
	}

	// At this point we have aparently reasonable cli options, create the logger that will be used in this main thread.
	logger := logging.New(defs.RuntimeLoggerPrefix, logging.Green)

	// Attempt to parse the url provided by the user - should be in full http://hostname:port format.
	apiHome, e := url.Parse(options.apiHome)

	if e != nil {
		logger.Errorf("invalid api (%s) host: %s", options.apiHome, e.Error())
		return
	}

	// Load the RSA private key file. This will be used to create the public key that will be sent to the server.
	key, e := security.ReadDeviceKeyFromFile(options.privateKeyfile)

	if e != nil {
		logger.Errorf("invalid file name: %s", e.Error())
		return
	}

	// Attempt to generate the shared secret that will be given to the server for encrypting digests to the device.
	sharedSecret, e := key.SharedSecret()

	if e != nil {
		logger.Errorf("invalid file name: %s", e.Error())
		return
	}

	var device beacon.Commandable
	// If the user has launched the application with the `-debug` flag, log to stdout rather than the blink1 device.
	if options.debugging {
		debugLog := logging.New(defs.DebugStateLoggerPrefix, logging.Cyan)
		device = &beacon.StateLogger{debugLog}
		logger.Debugf("shared secret: \n\n%s\n\n", sharedSecret)
	} else {
		var e error
		device, e = blink1.OpenNextDevice()

		if e != nil {
			logger.Errorf("unable to open blink device: %s", e.Error())
			return
		}
	}

	defer device.Close()
	defer device.SetState(blink1.State{})

	logger.Debugf("creating websocket subscriber w/ api: %s", apiHome.String())

	var subscriber beacon.Subscriber = &beacon.WebsocketSubscriber{
		Config: beacon.WebsocketConfig{
			APIHome: *apiHome,
			Secret:  sharedSecret,
		},
	}

	// If the user has provided a device name, pre-register the name with our shared secret before continuing.
	if options.deviceName != "" {
		e := subscriber.Preregister(options.deviceName)

		if e != nil {
			logger.Errorf("unable to register name \"%s\" with api: %s", options.deviceName, e.Error())
			return
		}
	}

	if e := subscriber.Connect(); e != nil {
		logger.Errorf("unable to open api subscription: %s (%s)", e.Error(), apiHome.String())
		return
	}

	defer subscriber.Close()

	commandStream := make(chan *bytes.Buffer, options.commandBuffer)
	feedbackStream := make(chan *beacon.Feedback, options.commandBuffer)

	bgSync := sync.WaitGroup{}
	delay, retries := time.Duration(int64(options.heartbeatDelay)*time.Second.Nanoseconds()), 0

	processors := []beacon.Processor{
		beacon.NewCommandProcessor(device, key, commandStream, feedbackStream),
		beacon.NewHeartbeatProcessor(subscriber, delay, uint(options.maxRetries)),
		beacon.NewFeedbackProcessor(feedbackStream, *apiHome),
	}

	// Iterate over each background processor, spawining each in a goroutine with a sync.WaitGroup.
	for _, p := range processors {
		bgSync.Add(1)
		go p.Start(&bgSync)
	}

	for subscriber.Connected() || retries < options.maxRetries {
		var e error

		if subscriber.Connected() {
			buffer := bytes.NewBuffer([]byte{})
			e = subscriber.ReadInto(buffer)

			// If there was no error, reset our retry counter, send the buffer into our stream and continue on.
			if e == nil {
				retries = 0
				commandStream <- buffer
				continue
			}
		}

		// If there was an error and we have reached our total count, break from the loop and log the error.
		if e != nil && retries+1 == options.maxRetries {
			logger.Errorf("max tries reached & bad read: %s", e.Error())
			break
		}

		if e != nil {
			// At this point, have received an error but we are able to continue on.
			logger.Warnf("bad read: %s, retrying after %d seconds", e.Error(), 5)
		}

		// Bump our retry count.
		retries++

		// Wait the specified amount of seconds.
		time.Sleep(time.Duration(time.Second.Nanoseconds() * int64(options.retryDelay)))

		logger.Infof("attempting retry: %d", retries)

		// If a device name was provided at startup, we need to pre-register again.
		if options.deviceName != "" {
			e := subscriber.Preregister(options.deviceName)

			// If we are unable to re-preregister (e.g the server is still down) continue on.
			if e != nil {
				logger.Warnf("failed preregister \"%s\" on retry: %s", options.deviceName, e.Error())
			}
		}

		// Retry our connection attempt at this point.
		subscriber.Connect()
	}

	// Close the command stream, terminating the command processor.
	close(commandStream)
	close(feedbackStream)

	// Wait for all background processors to complete.
	bgSync.Wait()
	logger.Warnf("connection loop terminated afte %d retries", retries)
}
