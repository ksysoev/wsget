package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"

	"github.com/TylerBrock/colorjson"
	"github.com/eiannone/keyboard"
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

type View struct {
	state string
}

func main() {
	fmt.Println("Connecting to", wsUrl, "...")
	wsInsp, err := NewWSInspector(wsUrl)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected")
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

	fmt.Println("to enter to request mode press ESC")
	for {
		if err := keyboard.Open(); err != nil {
			panic(err)
		}

		_, key, err := keyboard.GetSingleKey()
		if err != nil {
			panic(err)
		}

		if key == keyboard.KeyCtrlC || key == keyboard.KeyCtrlD {
			return
		}

		if key == keyboard.KeyEsc {
			keyboard.Close()
			fmt.Println("Request mode. Press Ctrl+S to send request. Press ESC to cancel request")

			keysEvents, err := keyboard.GetKeys(10)
			if err != nil {
				panic(err)
			}
			buffer := ""
			for {
				event := <-keysEvents
				if event.Err != nil {
					panic(event.Err)
				}

				if key == keyboard.KeyCtrlC || key == keyboard.KeyCtrlD {
					return
				}

				if event.Key == keyboard.KeyCtrlS {
					break
				}

				if event.Key == keyboard.KeyEsc {
					buffer = ""
					break
				}

				if event.Key == keyboard.KeySpace {
					fmt.Print(" ")
					buffer += " "
					continue
				}

				if event.Key == keyboard.KeyEnter {
					fmt.Print("\n")
					buffer += "\n"
					continue
				}

				if event.Key == keyboard.KeyBackspace || keyboard.KeyDelete == event.Key || event.Key == 127 {
					if len(buffer) == 0 {
						continue
					}

					if buffer[len(buffer)-1] == '\n' {
						continue
					}

					fmt.Print("\b \b")
					buffer = buffer[:len(buffer)-1]
					continue
				}

				if event.Key > 0 {
					continue
				}

				fmt.Print(string(event.Rune))

				buffer += string(event.Rune)
			}

			if buffer == "" {
				continue
			}

			err = wsInsp.Send(buffer)
			if err != nil {
				log.Fatal(err)
			}
		}
		keyboard.Close()
	}
}
