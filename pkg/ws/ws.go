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
	"github.com/fatih/color"
	"github.com/ksysoev/wsget/pkg/core"
)

const (
	wsMessageBufferSize   = 100
	headerPartsNumber     = 2
	dialTimeout           = 15 * time.Second
	defaultMaxMessageSize = 1024 * 1024
)

type Connection struct {
	url      *url.URL
	ws       *websocket.Conn
	messages chan core.Message
	opts     Options
	wg       sync.WaitGroup
	l        sync.Mutex
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

	wsInsp := &Connection{url: parsedURL, messages: make(chan core.Message, wsMessageBufferSize), opts: opts}

	return wsInsp, nil
}

func (c *Connection) Connect(ctx context.Context) error {
	c.l.Lock()
	defer c.l.Unlock()

	if c.ws != nil {
		return fmt.Errorf("connection already established")
	}

	httpCli := &http.Client{
		Transport: &requestLogger{
			transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: c.opts.SkipSSLVerification}, //nolint:gosec // Skip SSL verification
			},
			verbose: c.opts.Verbose,
		},
		Timeout: dialTimeout,
	}

	wsOpts := &websocket.DialOptions{
		HTTPClient: httpCli,
	}

	if len(c.opts.Headers) > 0 {
		Headers := make(http.Header)
		for _, headerInput := range c.opts.Headers {
			splited := strings.Split(headerInput, ":")
			if len(splited) != headerPartsNumber {
				return fmt.Errorf("invalid header: %s", headerInput)
			}

			header := strings.TrimSpace(splited[0])
			value := strings.TrimSpace(splited[1])

			Headers.Add(header, value)
		}

		wsOpts.HTTPHeader = Headers
	}

	ws, resp, err := websocket.Dial(ctx, c.url.String(), wsOpts)
	if err != nil {
		return err
	}

	if resp.Body != nil {
		resp.Body.Close()
	}

	ws.SetReadLimit(defaultMaxMessageSize)

	c.ws = ws

	c.wg.Add(1)

	go c.handleResponses(ctx, ws)

	return nil
}

// Messages returns a channel that receives messages from the WebSocket connection.
func (c *Connection) Messages() <-chan core.Message {
	return c.messages
}

// Hostname returns the hostname of the WebSocket server.
func (c *Connection) Hostname() string {
	return c.url.Hostname()
}

// handleResponses reads messages from the websocket connection and sends them to the Messages channel.
// It runs in a loop until the connection is closed or an error occurs.
func (c *Connection) handleResponses(ctx context.Context, ws *websocket.Conn) {
	defer func() {
		c.wg.Done()
		close(c.messages)
	}()

	for ctx.Err() == nil {
		msgType, reader, err := ws.Reader(ctx)
		if err != nil {
			c.handleError(err)
			return
		}

		if msgType == websocket.MessageBinary {
			c.handleError(fmt.Errorf("unexpected binary message"))
			return
		}

		data, err := io.ReadAll(reader)
		if err != nil {
			c.handleError(err)
			return
		}

		c.messages <- core.Message{Type: core.Response, Data: string(data)}
	}
}

func (c *Connection) handleError(err error) {
	if errors.Is(err, context.Canceled) || errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) {
		return
	}

	color.New(color.FgRed).Println("Fail read from connection:", err)
}

// Send sends a message to the websocket connection and returns a Message and an error.
// It takes a string message as input and returns a pointer to a Message struct and an error.
// The Message struct contains the message type and data.
func (c *Connection) Send(msg string) (*core.Message, error) {
	c.l.Lock()
	defer c.l.Unlock()

	if err := c.ws.Write(context.TODO(), websocket.MessageText, []byte(msg)); err != nil {
		return nil, err
	}

	return &core.Message{Type: core.Request, Data: msg}, nil
}

// Close closes the WebSocket connection.
// If the connection is already closed, it returns immediately.
func (c *Connection) Close() {
	c.l.Lock()
	defer c.l.Unlock()

	c.ws.Close(websocket.StatusNormalClosure, "closing connection")

	c.wg.Wait()
}
