package ws

import (
	"context"
	"crypto/tls"
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
	"github.com/ksysoev/wsget/pkg/core"
)

const (
	wsMessageBufferSize   = 100
	headerPartsNumber     = 2
	dialTimeout           = 15 * time.Second
	defaultMaxMessageSize = 1024 * 1024
)

type Connection struct {
	url       *url.URL
	ws        *websocket.Conn
	onMessage func(core.Message)
	opts      *websocket.DialOptions
	l         sync.Mutex
	ready     chan struct{}
}

type Options struct {
	Headers             []string
	SkipSSLVerification bool
	Verbose             bool
}

// New creates a new WebSocket connection to the specified URL with the given options.
// It returns a Connection object and an error if any occurred.
func New(wsURL string, opts Options) (*Connection, error) {
	parsedURL, err := url.Parse(wsURL)
	if err != nil {
		return nil, err
	}

	httpCli := &http.Client{
		Transport: &requestLogger{
			transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: opts.SkipSSLVerification}, //nolint:gosec // Skip SSL verification
			},
			verbose: opts.Verbose,
		},
		Timeout: dialTimeout,
	}

	wsOpts := &websocket.DialOptions{
		HTTPClient: httpCli,
	}

	if len(opts.Headers) > 0 {
		Headers := make(http.Header)
		for _, headerInput := range c.opts.Headers {
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

func (c *Connection) SetOnMessage(onMessage func(core.Message)) {
	c.l.Lock()
	defer c.l.Unlock()

	c.onMessage = onMessage
}

func (c *Connection) Connect(ctx context.Context) error {
	if c.onMessage == nil {
		return fmt.Errorf("onMessage callback is not set")
	}
	ws, resp, err := websocket.Dial(ctx, c.url.String(), c.opts)
	if err != nil {
		return err
	}

	if resp.Body != nil {
		resp.Body.Close()
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

// Hostname returns the hostname of the WebSocket server.
func (c *Connection) Hostname() string {
	return c.url.Hostname()
}

// handleResponses reads messages from the websocket connection and sends them to the Messages channel.
// It runs in a loop until the connection is closed or an error occurs.
func (c *Connection) handleResponses(ctx context.Context, ws *websocket.Conn) error {
	for ctx.Err() == nil {
		msgType, reader, err := ws.Reader(ctx)
		if err != nil {

			return c.handleError(err)
		}

		if msgType == websocket.MessageBinary {
			return c.handleError(fmt.Errorf("unexpected binary message"))
		}

		data, err := io.ReadAll(reader)
		if err != nil {
			return c.handleError(err)
		}

		c.onMessage(core.Message{Type: core.Response, Data: string(data)})
	}

	return nil
}

func (c *Connection) handleError(err error) error {
	if errors.Is(err, context.Canceled) || errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) {
		return nil
	}

	return fmt.Errorf("Fail read from connection: %w", err)
}

// Send sends a message to the websocket connection and returns a Message and an error.
// It takes a string message as input and returns a pointer to a Message struct and an error.
// The Message struct contains the message type and data.
func (c *Connection) Send(ctx context.Context, msg string) error {
	select {
	case <-c.ready:
	case <-ctx.Done():
		return ctx.Err()
	}

	return c.ws.Write(ctx, websocket.MessageText, []byte(msg))
}

// Close closes the WebSocket connection.
// If the connection is already closed, it returns immediately.
func (c *Connection) Close() error {
	select {
	case <-c.ready:
	default:
		return fmt.Errorf("connection is not established")
	}

	return c.ws.Close(websocket.StatusNormalClosure, "closing connection")
}
