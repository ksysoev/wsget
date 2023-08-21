package ws

import (
	"log"

	"golang.org/x/net/websocket"
)

type MessageType uint8

const (
	NotDefined MessageType = iota
	Request
	Response
)

type Message struct {
	Type MessageType `json:"type"`
	Data string      `json:"data"`
}

type WSConnection struct {
	ws       *websocket.Conn
	Messages chan Message
}

func NewWS(url string) (*WSConnection, error) {
	ws, err := websocket.Dial(url, "", "http://localhost")

	if err != nil {
		return nil, err
	}

	messages := make(chan Message, 100)

	go func(messages chan Message) {
		for {
			var msg string
			err = websocket.Message.Receive(ws, &msg)
			if err != nil {
				log.Fatal("Fail to read from WS connection:", err)
			}

			messages <- Message{Type: Response, Data: msg}
		}
	}(messages)

	return &WSConnection{ws: ws, Messages: messages}, nil
}

func (wsInsp *WSConnection) Send(msg string) error {
	err := websocket.Message.Send(wsInsp.ws, msg)

	if err != nil {
		return err
	}

	wsInsp.Messages <- Message{Type: Request, Data: msg}

	return nil
}

func (wsInsp *WSConnection) Close() {
	wsInsp.ws.Close()
}
