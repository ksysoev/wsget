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

	f := colorjson.NewFormatter()
	f.Indent = 2

	var obj any
	json.Unmarshal([]byte(msg), &obj)

	s, _ := f.Marshal(obj)

	fmt.Println(string(s))
}
