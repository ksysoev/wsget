package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/ksysoev/wsget/pkg/ws"

	"github.com/TylerBrock/colorjson"
	"github.com/eiannone/keyboard"
	"github.com/fatih/color"
)

var wsUrl string
var OutputFH *os.File
var InputFH *os.File

func init() {
	url := flag.String("u", "", "ws url")
	outputFile := flag.String("o", "", "output file")
	inputFile := flag.String("i", "", "input file")

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

	if inputFile != nil && *inputFile != "" {
		var err error
		InputFH, err = os.Open(*inputFile)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func main() {
	fmt.Println("Connecting to", wsUrl, "...")
	wsInsp, err := ws.NewWS(wsUrl)
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

	if InputFH != nil {
		go func() {
			scanner := bufio.NewScanner(InputFH)
			for scanner.Scan() {
				err = wsInsp.Send(scanner.Text())
				if err != nil {
					fmt.Println("Fail to send request:", err)
				}
			}
		}()
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

		case msg := <-wsInsp.Messages:
			var output []byte
			var obj any
			err = json.Unmarshal([]byte(msg.Data), &obj)
			var formater *colorjson.Formatter
			if err != nil {
				// Fail to parse Json just print as a string
				if msg.Type == ws.Request {
					formater = reqFormater
				} else {
					formater = respFormater
				}
				output = []byte(formater.KeyColor.Sprintf("%s", msg.Data))
			} else {
				// Parse Json and print with colors
				if msg.Type == ws.Request {
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

		switch e.Key {
		case keyboard.KeyCtrlC, keyboard.KeyCtrlD:
			return buffer, fmt.Errorf("interrupted")
		case keyboard.KeyCtrlS:
			if buffer == "" {
				return buffer, fmt.Errorf("cannot send empty request")
			}
			return buffer, nil
		case keyboard.KeyEsc:
			return "", nil

		case keyboard.KeySpace:
			fmt.Print(" ")
			buffer += " "
			continue

		case keyboard.KeyEnter:
			fmt.Print("\n")
			buffer += "\n"
			continue

		case keyboard.KeyBackspace, keyboard.KeyDelete, 127:
			if len(buffer) == 0 {
				continue
			}

			if buffer[len(buffer)-1] == '\n' {
				continue
			}

			fmt.Print("\b \b")
			buffer = buffer[:len(buffer)-1]
			continue
		default:
			if e.Key > 0 {
				continue
			}
			fmt.Print(string(e.Rune))
			buffer += string(e.Rune)
		}
	}

	return buffer, fmt.Errorf("keyboard stream was unexpectably closed")
}
