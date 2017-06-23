package main

import "io"
import "os"
import "log"
import "fmt"
import "flag"
import "sync"
import "time"
import "bytes"
import "net/http"

import "github.com/hink/go-blink1"
import "github.com/gorilla/websocket"
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
	}{}

	flag.StringVar(&options.apiHome, "api", "0.0.0.0:12345", "the hostname of the beacon.api server")
	flag.BoolVar(&options.debugging, "debug", false, "if true, the client will not attempt to open the blink device")
	flag.IntVar(&options.commandBuffer, "command-buffer", 2, "amount of allowed commands to buffer")
	flag.IntVar(&options.heartbeatDelay, "heartbeat-delay", 10, "amount of seconds between heartbeat pings")
	flag.StringVar(&options.privateKeyfile, "privte-key", ".keys/private.pem", "the filename of the private key")
	flag.Parse()

	if len(options.apiHome) < 1 {
		flag.PrintDefaults()
		return
	}

	logger := log.New(os.Stdout, "beacon client", log.Ldate|log.Ltime|log.Lshortfile)
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
		device = &beacon.StateLogger{log.New(os.Stdout, "beacon log device", log.Ldate|log.Ltime|log.Lshortfile)}
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
	dialer, apiUrl := websocket.Dialer{}, fmt.Sprintf("ws://%s/%s", options.apiHome, defs.APIRegistrationEndpoint)
	header := http.Header{}

	header.Set(defs.APIAuthorizationHeader, sharedSecret)

	logger.Printf("opening connection to api: %s", apiUrl)
	ws, _, e := dialer.Dial(apiUrl, header)

	if e != nil {
		logger.Fatalf("unable to connect to %s: %s", apiUrl, e.Error())
		return
	}

	defer ws.Close()
	commandStream, wait, connected := make(chan *bytes.Buffer, options.commandBuffer), sync.WaitGroup{}, true
	delay := time.Duration(int64(options.heartbeatDelay) * time.Second.Nanoseconds())

	commands := beacon.NewCommandProcessor(device, commandStream)
	heartbeat := beacon.NewHeartbeatProcessor(ws, delay)

	for _, p := range []beacon.Processor{commands, heartbeat} {
		wait.Add(1)
		go p.Start(&wait)
	}

	for connected {
		_, reader, e := ws.NextReader()

		if e != nil {
			logger.Printf("lost connection to server: %s", e.Error())
			connected = false
			break
		}

		buffer := bytes.NewBuffer([]byte{})

		if _, e := io.Copy(buffer, reader); e != nil {
			logger.Printf("unable to decode header from message: %s", e.Error())
			continue
		}

		commandStream <- buffer
	}
}
