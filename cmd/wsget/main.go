package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/TylerBrock/colorjson"
	"golang.org/x/net/websocket"
)

var wsUrl string

func init() {
	url := flag.String("u", "", "ws url")
	flag.Parse()

	if url == nil || *url == "" {
		log.Fatal("url is requered")
	}

	wsUrl = *url
}

type WSInspector struct {
	ws       *websocket.Conn
	messages chan string
}

func NewWSInspector(url string) (*WSInspector, error) {
	ws, err := websocket.Dial(url, "", "http://localhost")

	if err != nil {
		return nil, err
	}

	messages := make(chan string, 100)

	go func(messages chan string) {
		for {
			var msg string
			err = websocket.Message.Receive(ws, &msg)
			if err != nil {
				log.Fatal(err)
			}

			messages <- msg
		}
	}(messages)

	return &WSInspector{ws: ws, messages: messages}, nil
}

func (wsInsp *WSInspector) Send(msg string) error {
	err := websocket.Message.Send(wsInsp.ws, msg)

	if err != nil {
		return err
	}

	wsInsp.messages <- msg

	return nil
}

func (wsInsp *WSInspector) Close() {
	wsInsp.ws.Close()
}

func main() {
	wsInsp, err := NewWSInspector(wsUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer wsInsp.Close()

	go func() {
		f := colorjson.NewFormatter()
		f.Indent = 2
		for msg := range wsInsp.messages {
			var obj any
			json.Unmarshal([]byte(msg), &obj)
			s, _ := f.Marshal(obj)
			fmt.Println(string(s), "\n")
		}
		return
	}()

	for {
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')

		err = wsInsp.Send(input)

		if err != nil {
			log.Fatal(err)
		}
	}
}
