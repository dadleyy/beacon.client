package main

import "fmt"
import "log"
import "golang.org/x/net/websocket"

func main() {
	origin, url := "http://0.0.0.0", "ws://0.0.0.0:12345/echo"

	log.Println("Dialing connection...")
	ws, err := websocket.Dial(url, "", origin)

	if err != nil {
		log.Fatal(err)
		return
	}

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
			msg := make([]byte, 512)

			if _, e := ws.Read(msg); e != nil {
				break
			}

			receiver <- string(msg)
		}

		receiver <- ""
	}()

	for {
		log.Println("listing for input or response")

		select {
		case input := <-sender:
			log.Printf("input[%s]\n", input)

			if input == "" {
				log.Println("exited via input")
				break
			}

			if _, e := ws.Write([]byte(input)); e != nil {
				log.Fatal(e)
				break
			}
		case response := <-receiver:
			if response == "" {
				log.Println("exited via response")
				break
			}

			log.Printf("response[%s]\n", response)
		}
	}
}
