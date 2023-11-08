package ws

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/fatih/color"
	"golang.org/x/net/websocket"
)

type MessageType uint8

const (
	NotDefined MessageType = iota
	Request
	Response
)

func (mt MessageType) String() string {
	switch mt {
	case Request:
		return "Request"
	case Response:
		return "Response"
	default:
		return "Not defined"
	}
}

const (
	WSMessageBufferSize = 100
	HeaderPartsNumber   = 2
)

type Message struct {
	Data string      `json:"data"`
	Type MessageType `json:"type"`
}

type Connection struct {
	ws        *websocket.Conn
	Messages  chan Message
	waitGroup *sync.WaitGroup
}

type Options struct {
	Headers             []string
	SkipSSLVerification bool
}

func NewWS(url string, opts Options) (*Connection, error) {
	cfg, err := websocket.NewConfig(url, "http://localhost")
	if err != nil {
		return nil, err
	}

	// This option could be useful for testing and development purposes.
	// Default value is false.
	// #nosec G402
	tlsConfig := &tls.Config{
		InsecureSkipVerify: opts.SkipSSLVerification,
	}
	cfg.TlsConfig = tlsConfig

	if len(opts.Headers) > 0 {
		Headers := make(http.Header)
		for _, headerInput := range opts.Headers {
			splited := strings.Split(headerInput, ":")
			if len(splited) != HeaderPartsNumber {
				return nil, fmt.Errorf("invalid header: %s", headerInput)
			}

			header := strings.TrimSpace(splited[0])
			value := strings.TrimSpace(splited[1])

			Headers.Add(header, value)
		}

		cfg.Header = Headers
	}

	ws, err := websocket.DialConfig(cfg)

	if err != nil {
		return nil, err
	}

	var waitGroup sync.WaitGroup

	messages := make(chan Message, WSMessageBufferSize)

	go func() {
		defer func() {
			waitGroup.Wait()
			close(messages)
		}()

		for {
			var msg string

			err = websocket.Message.Receive(ws, &msg)
			if err != nil {
				if err.Error() == "EOF" {
					color.New(color.FgRed).Println("Connection closed by the server")
				} else {
					color.New(color.FgRed).Println("Fail read from connection: ", err)
				}

				return
			}

			messages <- Message{Type: Response, Data: msg}
		}
	}()

	return &Connection{ws: ws, Messages: messages, waitGroup: &waitGroup}, nil
}

func (wsInsp *Connection) Send(msg string) (*Message, error) {
	wsInsp.waitGroup.Add(1)
	defer wsInsp.waitGroup.Done()

	err := websocket.Message.Send(wsInsp.ws, msg)

	if err != nil {
		return nil, err
	}

	return &Message{Type: Request, Data: msg}, nil
}

func (wsInsp *Connection) Close() {
	wsInsp.ws.Close()
}
