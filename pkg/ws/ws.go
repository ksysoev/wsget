package ws

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/coder/websocket"
)

const (
	headerPartsNumber     = 2
	dialTimeout           = 15 * time.Second
	defaultMaxMessageSize = 1024 * 1024
)

type reader interface {
	Read(p []byte) (n int, err error)
}

type Connection struct {
	url       *url.URL
	ws        *websocket.Conn
	onMessage func(context.Context, []byte)
	opts      *websocket.DialOptions
	ready     chan struct{}
	l         sync.Mutex
}

type Options struct {
	Output              io.Writer
	Headers             []string
	SkipSSLVerification bool
}

// New initializes a new WebSocket connection configuration with specified URL and options.
// It takes wsURL, a string representing the WebSocket URL, and opts, an instance of Options with custom settings.
// It returns a pointer to a Connection and possible error if the URL is empty, poorly formatted, or headers are invalid.
func New(wsURL string, opts Options) (*Connection, error) {
	if wsURL == "" {
		return nil, errors.New("url is empty")
	}

	parsedURL, err := url.Parse(wsURL)
	if err != nil {
		return nil, err
	}

	httpCli := &http.Client{
		Transport: newRequestLogger(opts.Output, opts.SkipSSLVerification),
		Timeout:   dialTimeout,
	}

	wsOpts := &websocket.DialOptions{
		HTTPClient: httpCli,
	}

	if len(opts.Headers) > 0 {
		Headers := make(http.Header)
		for _, headerInput := range opts.Headers {
			splited := strings.Split(headerInput, ":")
			if len(splited) != headerPartsNumber {
				return nil, fmt.Errorf("invalid header: %s", headerInput)
			}

			header := strings.TrimSpace(splited[0])
			value := strings.TrimSpace(splited[1])

			Headers.Add(header, value)
		}

		wsOpts.HTTPHeader = Headers
	}

	return &Connection{
		url:   parsedURL,
		opts:  wsOpts,
		ready: make(chan struct{}),
	}, nil
}

// SetOnMessage sets the callback function to handle incoming messages on the connection.
// It takes onMessage, a function with parameters context.Context and a byte slice [], as input.
// The method does not return any value and is thread-safe, locking access to the callback function.
func (c *Connection) SetOnMessage(onMessage func(context.Context, []byte)) {
	c.l.Lock()
	defer c.l.Unlock()

	c.onMessage = onMessage
}

// Connect establishes a WebSocket connection using the specified context.
// It returns an error if the onMessage callback is not set, the connection attempt fails,
// or if a connection is already established.
// The method locks the connection during setup to ensure thread safety and sets a default read limit on the WebSocket.
func (c *Connection) Connect(ctx context.Context) error {
	if c.onMessage == nil {
		return fmt.Errorf("onMessage callback is not set")
	}

	ws, resp, err := websocket.Dial(ctx, c.url.String(), c.opts)
	if err != nil {
		return c.handleError(err)
	}

	if resp.Body != nil {
		_ = resp.Body.Close()
	}

	c.l.Lock()
	if c.ws != nil {
		c.l.Unlock()
		return fmt.Errorf("connection already established")
	}

	c.ws = ws
	close(c.ready)

	c.l.Unlock()

	ws.SetReadLimit(defaultMaxMessageSize)

	return c.handleResponses(ctx, ws)
}

// Hostname retrieves the host name part of the URL stored in the Connection struct.
// It returns a string representing the host name.
func (c *Connection) Hostname() string {
	return c.url.Hostname()
}

// handleResponses manages incoming messages on a WebSocket connection until the context is canceled.
// It takes a context (ctx) for cancellation control and a websocket connection (ws) for message communication.
// It returns an error if there is an issue reading from the WebSocket or if handling a message fails.
// The function terminates without error if the context is canceled.
func (c *Connection) handleResponses(ctx context.Context, ws *websocket.Conn) error {
	for ctx.Err() == nil {
		msgType, reader, err := ws.Reader(ctx)
		if err != nil {
			return c.handleError(err)
		}

		if err := c.handleMessage(ctx, msgType, reader); err != nil {
			return c.handleError(err)
		}
	}

	return nil
}

// handleMessage processes an incoming WebSocket message for the Connection.
// It takes ctx of type context.Context, msgType of type websocket.MessageType, and msgReader of type reader.
// It returns an error if the message type is binary or if reading from the reader fails.
// The function reads all data from msgReader and invokes the onMessage callback with the read data.
func (c *Connection) handleMessage(ctx context.Context, msgType websocket.MessageType, msgReader reader) error {
	if msgType == websocket.MessageBinary {
		return fmt.Errorf("unexpected binary message")
	}

	data, err := io.ReadAll(msgReader)
	if err != nil {
		return fmt.Errorf("fail to read message: %w", err)
	}

	c.onMessage(ctx, data)

	return nil
}

// handleError processes an error arising from a WebSocket connection.
// It takes an err parameter of type error and returns an error value.
// The method returns nil if the error is context.Canceled, io.EOF, net.ErrClosed or a websocket.StatusNormalClosure.
// It returns a formatted error message if the error is a websocket.CloseError with any other close code.
func (c *Connection) handleError(err error) error {
	if errors.Is(err, context.Canceled) || errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) {
		return nil
	}

	var ce websocket.CloseError
	if errors.As(err, &ce) {
		if ce.Code == websocket.StatusNormalClosure {
			return nil
		}

		return fmt.Errorf("connection closed: %s %s", ce.Code, ce.Reason)
	}

	return fmt.Errorf("fail read from connection: %w", err)
}

// Send transmits a message over an established WebSocket connection within a given context.
// It takes ctx of type context.Context and msg of type string as parameters.
// It returns an error if the context is canceled or if there is a failure writing to the WebSocket.
// The function waits for the connection to be ready before sending the message.
func (c *Connection) Send(ctx context.Context, msg string) error {
	select {
	case <-c.ready:
	case <-ctx.Done():
		return ctx.Err()
	}

	return c.ws.Write(ctx, websocket.MessageText, []byte(msg))
}

// Close shuts down an established WebSocket connection gracefully.
// It returns an error if the connection is not yet established.
// The function ensures a normal closure status is sent to the WebSocket server.
func (c *Connection) Close() error {
	select {
	case <-c.ready:
	default:
		return fmt.Errorf("connection is not established")
	}

	return c.ws.Close(websocket.StatusNormalClosure, "closing connection")
}

// Ready returns a channel that is closed when the WebSocket connection is established.
func (c *Connection) Ready() <-chan struct{} {
	return c.ready
}
