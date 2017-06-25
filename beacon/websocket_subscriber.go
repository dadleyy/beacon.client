package beacon

import "io"
import "fmt"
import "bytes"
import "net/url"
import "net/http"
import "encoding/json"
import "github.com/gorilla/websocket"
import "github.com/dadleyy/beacon.client/beacon/defs"

// WebsocketConfig holds the necessary information to subscribe to the api via websocket
type WebsocketConfig struct {
	APIHome *url.URL
	Secret  string
}

// WebsocketSubscriber is a websocket implementation of the Subscriber interface
type WebsocketSubscriber struct {
	Config     WebsocketConfig
	connection *websocket.Conn
	connected  uint
}

// Preregister attempts to reserve the provided device name w/ the server
func (subscriber *WebsocketSubscriber) Preregister(name string) error {
	request := struct {
		Name   string `json:"name"`
		Secret string `json:"shared_secret"`
	}{name, subscriber.Config.Secret}
	buf, e := json.Marshal(&request)

	if e != nil {
		return e
	}

	response, e := http.Post(subscriber.registrationAddress(), "application/json", bytes.NewBuffer(buf))

	if e != nil {
		return e
	}

	if response.StatusCode != 200 {
		return fmt.Errorf("invalid-response")
	}

	return nil
}

// Ping simply writes the data to the websocket
func (subscriber *WebsocketSubscriber) Ping(data []byte) error {
	writer, e := subscriber.connection.NextWriter(websocket.TextMessage)

	if e != nil {
		return e
	}

	defer writer.Close()

	_, e = writer.Write(data)

	return e
}

// ReadInto opens a new reader from the websocket and copies the data into the writer
func (subscriber *WebsocketSubscriber) ReadInto(writer io.Writer) error {
	_, r, e := subscriber.connection.NextReader()

	if e != nil {
		subscriber.connected = 0
		return e
	}

	_, e = io.Copy(writer, r)

	return e
}

// Connected returns true while the websocket is open
func (subscriber *WebsocketSubscriber) Connected() bool {
	return subscriber.connected == 1
}

// Close closes the websocket connection
func (subscriber *WebsocketSubscriber) Close() error {
	subscriber.connected = 0
	return subscriber.connection.Close()
}

// Connect opens the websocket connection
func (subscriber *WebsocketSubscriber) Connect() (e error) {
	config, header, dialer := subscriber.Config, http.Header{}, websocket.Dialer{}
	header.Set(defs.APIAuthorizationHeader, config.Secret)
	subscriber.connection, _, e = dialer.Dial(subscriber.websocketAddress(), header)

	if e == nil {
		subscriber.connected = 1
	}

	return
}

func (subscriber *WebsocketSubscriber) websocketAddress() string {
	u, e := url.Parse(subscriber.Config.APIHome.String())

	if e != nil {
		return ""
	}

	u.Scheme = "ws"
	return u.String()
}

func (subscriber *WebsocketSubscriber) registrationAddress() string {
	return subscriber.Config.APIHome.String()
}
