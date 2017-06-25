package main

import "os"
import "log"
import "flag"
import "sync"
import "time"
import "bytes"
import "net/url"

import "github.com/hink/go-blink1"
import "github.com/dadleyy/beacon.client/beacon"
import "github.com/dadleyy/beacon.client/beacon/defs"
import "github.com/dadleyy/beacon.client/beacon/security"

func main() {
	options := struct {
		apiHome        string
		debugging      bool
		commandBuffer  int
		heartbeatDelay int
		privateKeyfile string
		deviceName     string
	}{}

	flag.StringVar(&options.apiHome, "api", "http://0.0.0.0:12345", "the hostname of the beacon.api server")
	flag.BoolVar(&options.debugging, "debug", false, "if true, the client will not attempt to open the blink device")
	flag.IntVar(&options.commandBuffer, "command-buffer", 2, "amount of allowed commands to buffer")
	flag.IntVar(&options.heartbeatDelay, "heartbeat-delay", 10, "amount of seconds between heartbeat pings")
	flag.StringVar(&options.privateKeyfile, "privte-key", ".keys/private.pem", "the filename of the private key")
	flag.StringVar(&options.deviceName, "device-name", "", "if provided, this will attempt to register the name with the api")
	flag.Parse()

	if len(options.apiHome) < 1 {
		flag.PrintDefaults()
		return
	}

	logger := log.New(os.Stdout, defs.RuntimeLoggerPrefix, defs.DefaultLogFlags)

	var device beacon.Commandable

	key, e := security.ReadDeviceKeyFromFile(options.privateKeyfile)

	if e != nil {
		logger.Printf("invalid file name: %s", e.Error())
		return
	}

	sharedSecret, e := key.SharedSecret()

	if e != nil {
		logger.Printf("invalid file name: %s", e.Error())
		return
	}

	if options.debugging {
		debugLog := log.New(os.Stdout, defs.DebugStateLoggerPrefix, defs.DefaultLogFlags)
		device = &beacon.StateLogger{debugLog}
		logger.Printf("shared secret: \n\n%s\n\n", sharedSecret)
	} else {
		var e error
		device, e = blink1.OpenNextDevice()

		if e != nil {
			logger.Fatalf("unable to open blink device: %s", e.Error())
			return
		}
	}

	defer device.Close()
	defer device.SetState(blink1.State{})

	apiHome, e := url.Parse(options.apiHome)

	if e != nil {
		logger.Fatalf("unable to open blink device: %s", e.Error())
		return
	}

	apiHome.Path = defs.APIRegistrationEndpoint

	websocket := &beacon.WebsocketSubscriber{
		Config: beacon.WebsocketConfig{
			APIHome: apiHome,
			Secret:  sharedSecret,
		},
	}

	if options.deviceName != "" {
		e := websocket.Preregister(options.deviceName)

		if e != nil {
			logger.Printf("unable to register name \"%s\" with api: %s", options.deviceName, e.Error())
			return
		}
	}

	if e := websocket.Connect(); e != nil {
		logger.Fatalf("unable to open api subscription: %s (%s)", e.Error(), apiHome.String())
		return
	}

	defer websocket.Close()
	commandStream, wait := make(chan *bytes.Buffer, options.commandBuffer), sync.WaitGroup{}
	delay := time.Duration(int64(options.heartbeatDelay) * time.Second.Nanoseconds())

	commands := beacon.NewCommandProcessor(device, commandStream)
	heartbeat := beacon.NewHeartbeatProcessor(websocket, delay)

	for _, p := range []beacon.Processor{commands, heartbeat} {
		wait.Add(1)
		go p.Start(&wait)
	}

	for websocket.Connected() {
		buffer := bytes.NewBuffer([]byte{})

		if e := websocket.ReadInto(buffer); e != nil {
			logger.Printf("unable to decode header from message: %s", e.Error())
			continue
		}

		commandStream <- buffer
	}
}
