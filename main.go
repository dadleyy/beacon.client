package main

import "io"
import "os"
import "log"
import "fmt"
import "flag"
import "sync"
import "bytes"

import "github.com/hink/go-blink1"
import "github.com/gorilla/websocket"
import "github.com/dadleyy/beacon.client/beacon"
import "github.com/dadleyy/beacon.client/beacon/defs"

func main() {
	options := struct {
		apiHome       string
		debugging     bool
		commandBuffer int
	}{}

	flag.StringVar(&options.apiHome, "api", "0.0.0.0:12345", "the hostname of the beacon.api server")
	flag.BoolVar(&options.debugging, "debug", false, "if true, the client will not attempt to open the blink device")
	flag.IntVar(&options.commandBuffer, "command-buffer", 2, "amount of allowed commands to buffer")
	flag.Parse()

	if len(options.apiHome) < 1 {
		flag.PrintDefaults()
		return
	}

	logger := log.New(os.Stdout, "beacon client", log.Ldate|log.Ltime|log.Lshortfile)
	var device beacon.Commandable

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
	dialer, endpoint := websocket.Dialer{}, fmt.Sprintf("ws://%s/%s", options.apiHome, defs.APIRegistrationEndpoint)

	logger.Printf("opening connection to api: %s", endpoint)
	ws, _, e := dialer.Dial(endpoint, nil)

	if e != nil {
		logger.Fatalf("unable to connect to %s: %s", endpoint, e.Error())
		return
	}

	defer ws.Close()
	commandStream, wait, connected := make(chan *bytes.Buffer, options.commandBuffer), sync.WaitGroup{}, true

	processor := beacon.NewCommandProcessor(device, commandStream)

	wait.Add(1)
	go processor.Start(&wait)

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
