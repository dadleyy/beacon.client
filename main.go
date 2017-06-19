package main

import "fmt"
import "log"
import "time"
import "encoding/json"

import "github.com/hink/go-blink1"
import "github.com/gorilla/websocket"

type ControlMessage struct {
	DeviceId string        `json:"device_id"`
	Red      uint8         `json:"red"`
	Blue     uint8         `json:"blue"`
	Green    uint8         `json:"green"`
	LED      uint8         `json:"led"`
	FadeTime time.Duration `json:"fade_time"`
	Duration time.Duration `json:"duration"`
}

func (message *ControlMessage) State() blink1.State {
	return blink1.State{
		Red:      message.Red,
		Green:    message.Green,
		Blue:     message.Blue,
		FadeTime: message.FadeTime,
		Duration: message.Duration,
	}
}

func main() {
	url := "ws://0.0.0.0:12345/register"

	device, e := blink1.OpenNextDevice()

	if e != nil {
		log.Printf("unable to open device: %s", e.Error())
	}

	defer device.Close()
	defer device.SetState(blink1.State{})

	log.Println("Dialing connection...")
	dialer := websocket.Dialer{}
	ws, _, err := dialer.Dial(url, nil)

	if err != nil {
		log.Fatal(err)
		return
	}

	defer ws.Close()

	sender, receiver := make(chan string), make(chan string)

	go func() {
		for {
			var s string

			if _, e := fmt.Scanln(&s); e != nil {
				break
			}

			sender <- s
		}

		sender <- ""
	}()

	go func() {
		for {
			log.Printf("reading from server...")
			_, reader, e := ws.NextReader()

			if e != nil {
				log.Printf("unable to decode state: %s", e.Error())
				break
			}

			decoder, message := json.NewDecoder(reader), ControlMessage{}

			if e := decoder.Decode(&message); e != nil {
				log.Printf("unable to decode state: %s", e.Error())
				break
			}

			log.Printf("recieved message, device[%s]: %v", message.DeviceId, message.State())

			if e := device.SetState(message.State()); e != nil {
				log.Printf("unable to decode state: %s", e.Error())
				break
			}
		}

		receiver <- ""
	}()

	connected := true

	for connected {
		log.Println("listing for input or response")

		select {
		case input := <-sender:
			log.Printf("input[%s]\n", input)

			if input == "" || input == "quit" {
				connected = false
				log.Println("exited via input")
				break
			}

			writer, e := ws.NextWriter(websocket.TextMessage)

			if e != nil {
				connected = false
				log.Fatal(e)
				break
			}

			if _, e := writer.Write([]byte(input)); e != nil {
				connected = false
				log.Fatal(e)
				break
			}
		case response := <-receiver:
			if response == "" {
				log.Println("exited via response")
				connected = false
				break
			}

			log.Printf("response[%s]\n", response)
		}
	}
}
