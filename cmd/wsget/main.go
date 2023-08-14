package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"

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

func main() {
	ws, err := websocket.Dial(wsUrl, "", "http://localhost")

	if err != nil {
		log.Fatal(err)
	}

	defer ws.Close()

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')

	err = websocket.Message.Send(ws, input)

	if err != nil {
		log.Fatal(err)
	}

	var msg string
	err = websocket.Message.Receive(ws, &msg)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(msg)
}
