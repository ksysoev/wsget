package ws

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"
	"github.com/fatih/color"
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
	DialTimeout         = 15 * time.Second
)

type Message struct {
	Data string      `json:"data"`
	Type MessageType `json:"type"`
}

type Connection struct {
	ws        *websocket.Conn
	messages  chan Message
	waitGroup *sync.WaitGroup
	hostname  string
	isClosed  atomic.Bool
}

type Options struct {
	Headers             []string
	SkipSSLVerification bool
}

type ConnectionHandler interface {
	Messages() <-chan Message
	Hostname() string
	Send(msg string) (*Message, error)
	Close()
}

// NewWS creates a new WebSocket connection to the specified URL with the given options.
// It returns a Connection object and an error if any occurred.
func NewWS(wsURL string, opts Options) (*Connection, error) {
	parsedURL, err := url.Parse(wsURL)
	if err != nil {
		return nil, err
	}

	httpCli := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: opts.SkipSSLVerification}, //nolint:gosec // Skip SSL verification
		},
		Timeout: DialTimeout,
	}

	wsOpts := &websocket.DialOptions{
		HTTPClient: httpCli,
	}

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

		wsOpts.HTTPHeader = Headers
	}

	ws, resp, err := websocket.Dial(context.TODO(), wsURL, wsOpts)
	if err != nil {
		return nil, err
	}

	if resp.Body != nil {
		resp.Body.Close()
	}

	var waitGroup sync.WaitGroup

	messages := make(chan Message, WSMessageBufferSize)

	wsInsp := &Connection{ws: ws, messages: messages, waitGroup: &waitGroup, hostname: parsedURL.Hostname()}

	go wsInsp.handleResponses()

	return wsInsp, nil
}

// Messages returns a channel that receives messages from the WebSocket connection.
func (wsInsp *Connection) Messages() <-chan Message {
	return wsInsp.messages
}

// Hostname returns the hostname of the WebSocket server.
func (wsInsp *Connection) Hostname() string {
	return wsInsp.hostname
}

// handleResponses reads messages from the websocket connection and sends them to the Messages channel.
// It runs in a loop until the connection is closed or an error occurs.
func (wsInsp *Connection) handleResponses() {
	defer func() {
		wsInsp.waitGroup.Wait()
		close(wsInsp.messages)
	}()

	for {
		msgType, reader, err := wsInsp.ws.Reader(context.TODO())
		if err != nil {
			wsInsp.handleError(err)
			return
		}

		if msgType == websocket.MessageBinary {
			wsInsp.handleError(fmt.Errorf("unexpected binary message"))
			return
		}

		data, err := io.ReadAll(reader)
		if err != nil {
			wsInsp.handleError(err)
			return
		}

		wsInsp.messages <- Message{Type: Response, Data: string(data)}
	}
}

func (wsInsp *Connection) handleError(err error) {
	if wsInsp.isClosed.Load() {
		return
	}

	if err.Error() == "EOF" {
		color.New(color.FgRed).Println("Connection closed by the server")
		return
	}

	color.New(color.FgRed).Println("Fail read from connection:", err)
}

// Send sends a message to the websocket connection and returns a Message and an error.
// It takes a string message as input and returns a pointer to a Message struct and an error.
// The Message struct contains the message type and data.
func (wsInsp *Connection) Send(msg string) (*Message, error) {
	wsInsp.waitGroup.Add(1)
	defer wsInsp.waitGroup.Done()

	if err := wsInsp.ws.Write(context.TODO(), websocket.MessageText, []byte(msg)); err != nil {
		return nil, err
	}

	return &Message{Type: Request, Data: msg}, nil
}

// Close closes the WebSocket connection.
// If the connection is already closed, it returns immediately.
func (wsInsp *Connection) Close() {
	if wsInsp.isClosed.Load() {
		return
	}

	wsInsp.isClosed.Store(true)

	wsInsp.ws.Close(websocket.StatusNormalClosure, "closing connection")
}
