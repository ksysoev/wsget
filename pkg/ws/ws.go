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
	"syscall"
	"time"

	"github.com/coder/websocket"
)

const (
	headerPartsNumber     = 2
	DefaultMaxMessageSize = 1024 * 1024
)

var ErrConnectionClosed = errors.New("connection closed")

type reader interface {
	Read(p []byte) (n int, err error)
}

type Connection struct {
	output    io.Writer
	url       *url.URL
	ws        *websocket.Conn
	onMessage func(context.Context, []byte)
	opts      *websocket.DialOptions
	ready     chan struct{}
	msgSize   int64
	l         sync.Mutex
}

type Options struct {
	Output              io.Writer
	Headers             []string
	SkipSSLVerification bool
	MaxMessageSize      int64
	Timeout             time.Duration
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
		return nil, fmt.Errorf("failed to parse WebSocket URL %q: %w", wsURL, err)
	}

	httpCli := &http.Client{
		Transport: newRequestLogger(opts.Output, opts.SkipSSLVerification),
		Timeout:   opts.Timeout,
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

	var msgSize int64 = DefaultMaxMessageSize

	return &Connection{
		url:     parsedURL,
		opts:    wsOpts,
		ready:   make(chan struct{}),
		msgSize: msgSize,
		output:  opts.Output,
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

	startTime := time.Now()
	ws, resp, err := websocket.Dial(ctx, c.url.String(), c.opts)

	if c.output != nil {
		handshakeDuration := time.Since(startTime)

		fmt.Fprintf(c.output, "WebSocket handshake completed in %v\n", handshakeDuration)
	}

	if err != nil {
		return fmt.Errorf("failed to dial WebSocket: %w", handleError(err))
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

	ws.SetReadLimit(c.msgSize)

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
			return fmt.Errorf("failed to read from WebSocket: %w", handleError(err))
		}

		if err := c.handleMessage(ctx, msgType, reader); err != nil {
			return fmt.Errorf("failed to handle message: %w", handleError(err))
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
func handleError(err error) error {
	if err == nil || errors.Is(err, context.Canceled) {
		return nil
	}

	if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) || errors.Is(err, syscall.EPIPE) {
		return ErrConnectionClosed
	}

	var ce websocket.CloseError
	if errors.As(err, &ce) {
		if ce.Code == websocket.StatusNormalClosure {
			return ErrConnectionClosed
		}

		return fmt.Errorf("connection closed: %s %s", ce.Code, ce.Reason)
	}

	return fmt.Errorf("connection error: %w", err)
}

// Send transmits a message over an established WebSocket connection within a given context.
// It takes ctx of type context.Context and msg of type string as parameters.
// It returns an error if the context is canceled or if there is a failure writing to the WebSocket.
// The function waits for the connection to be ready before sending the message.
func (c *Connection) Send(ctx context.Context, msg string) error {
	select {
	case <-c.ready:
	case <-ctx.Done():
		return fmt.Errorf("context canceled while waiting to send: %w", ctx.Err())
	}

	err := c.ws.Write(ctx, websocket.MessageText, []byte(msg))
	if err != nil {
		return fmt.Errorf("failed to write to WebSocket: %w", handleError(err))
	}

	return nil
}

// Ping sends a ping frame to the WebSocket server to check the connection's liveness.
// It takes ctx of type context.Context as a parameter.
// It returns an error if the context is canceled or if there is a failure sending the ping
func (c *Connection) Ping(ctx context.Context) error {
	select {
	case <-c.ready:
	case <-ctx.Done():
		return fmt.Errorf("context canceled while waiting to ping: %w", ctx.Err())
	}

	err := c.ws.Ping(ctx)
	if err != nil {
		return fmt.Errorf("failed to send ping: %w", handleError(err))
	}

	return nil
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
