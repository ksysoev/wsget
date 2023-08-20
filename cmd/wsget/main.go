package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/TylerBrock/colorjson"
	"github.com/eiannone/keyboard"
	"github.com/fatih/color"
	"golang.org/x/net/websocket"
)

var wsUrl string
var OutputFH *os.File

func init() {
	url := flag.String("u", "", "ws url")
	outputFile := flag.String("o", "", "output file")

	flag.Parse()

	if url == nil || *url == "" {
		log.Fatal("url is requered")
	}

	wsUrl = *url

	if outputFile != nil && *outputFile != "" {
		var err error
		OutputFH, err = os.Create(*outputFile)
		if err != nil {
			log.Fatal(err)
		}
	}
}

type Message struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

type WSInspector struct {
	ws       *websocket.Conn
	messages chan Message
}

func NewWSInspector(url string) (*WSInspector, error) {
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

			messages <- Message{Type: "response", Data: msg}
		}
	}(messages)

	return &WSInspector{ws: ws, messages: messages}, nil
}

func (wsInsp *WSInspector) Send(msg string) error {
	err := websocket.Message.Send(wsInsp.ws, msg)

	if err != nil {
		return err
	}

	wsInsp.messages <- Message{Type: "request", Data: msg}

	return nil
}

func (wsInsp *WSInspector) Close() {
	wsInsp.ws.Close()
}

func main() {
	fmt.Println("Connecting to", wsUrl, "...")
	wsInsp, err := NewWSInspector(wsUrl)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected")
	defer wsInsp.Close()

	reqFormater := colorjson.NewFormatter()
	reqFormater.Indent = 2
	reqFormater.KeyColor = color.New(color.FgGreen, color.Bold)

	respFormater := colorjson.NewFormatter()
	respFormater.Indent = 2
	respFormater.KeyColor = color.New(color.FgHiRed, color.Bold)

	if err := keyboard.Open(); err != nil {
		panic(err)
	}
	defer keyboard.Close()

	keysEvents, err := keyboard.GetKeys(10)
	if err != nil {
		panic(err)
	}

	fmt.Println("Connection Mode: Press ESC to enter Request mode")
	for {
		select {
		case event := <-keysEvents:
			switch event.Key {
			case keyboard.KeyCtrlC, keyboard.KeyCtrlD:
				return

			case keyboard.KeyEsc:
				fmt.Println("Request Mode: Type your API request and press Ctrl+S to send it. Press ESC to cancel request")
				req, err := requestMode(keysEvents)

				if err != nil {
					if err.Error() == "interrupted" {
						return
					}

					fmt.Println(err)
				}

				if req != "" {
					err = wsInsp.Send(req)
					if err != nil {
						fmt.Println("Fail to send request:", err)
					}
				}

				fmt.Println("Connection Mode: Press ESC to enter Request mode")
			}

		case msg := <-wsInsp.messages:
			var output []byte
			var obj any
			err = json.Unmarshal([]byte(msg.Data), &obj)
			var formater *colorjson.Formatter
			if err != nil {
				// Fail to parse Json just print as a string
				if msg.Type == "request" {
					formater = reqFormater
				} else {
					formater = respFormater
				}
				output = []byte(formater.KeyColor.Sprintf("%s", msg.Data))
			} else {
				// Parse Json and print with colors
				if msg.Type == "request" {
					formater = reqFormater
				} else {
					formater = respFormater
				}

				output, err = formater.Marshal(obj)
				if err != nil {
					log.Fatalln("Fail to format JSON: ", err, msg)
				}
			}
			fmt.Printf("%s\n\n", string(output))

			if OutputFH != nil {
				fmt.Fprintln(OutputFH, string(output))
			}
		}
	}
}

func requestMode(keyStream <-chan keyboard.KeyEvent) (string, error) {
	buffer := ""
	for e := range keyStream {
		if e.Err != nil {
			return buffer, e.Err
		}

		switch {
		case e.Key == keyboard.KeyCtrlC, e.Key == keyboard.KeyCtrlD:
			return buffer, fmt.Errorf("interrupted")
		case e.Key == keyboard.KeyCtrlS:
			if buffer == "" {
				return buffer, fmt.Errorf("cannot send empty request")
			}
			return buffer, nil
		case e.Key == keyboard.KeyEsc:
			return "", nil

		case e.Key == keyboard.KeySpace:
			fmt.Print(" ")
			buffer += " "
			continue

		case e.Key == keyboard.KeyEnter:
			fmt.Print("\n")
			buffer += "\n"
			continue

		case e.Key == keyboard.KeyBackspace, e.Key == keyboard.KeyDelete, e.Key == 127:
			if len(buffer) == 0 {
				continue
			}

			if buffer[len(buffer)-1] == '\n' {
				continue
			}

			fmt.Print("\b \b")
			buffer = buffer[:len(buffer)-1]
			continue
		case e.Key > 0:
			// Ignore rest of special keys
			continue
		default:
			fmt.Print(string(e.Rune))
			buffer += string(e.Rune)
		}
	}

	return buffer, fmt.Errorf("keyboard stream was unexpectably closed")
}
