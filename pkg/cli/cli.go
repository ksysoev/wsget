package cli

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"strings"

	"github.com/eiannone/keyboard"
	"github.com/ksysoev/wsget/pkg/formater"
	"github.com/ksysoev/wsget/pkg/ws"
)

const (
	LINE_UP    = "\033[1A"
	LINE_CLEAR = "\x1b[2K"

	HISTORY_FILE  = ".wsget_history"
	HISTORY_LIMIT = 100

	MACOS_DELETE_KEY = 127

	KEYBOARD_BUFFER_SIZE = 10
)

type CLI struct {
	formater *formater.Formater
	history  *History
	wsConn   *ws.WSConnection
}

func NewCLI(wsConn *ws.WSConnection) *CLI {
	currentUser, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	homeDir := currentUser.HomeDir

	return &CLI{
		formater: formater.NewFormatter(),
		history:  NewHistory(homeDir+"/"+HISTORY_FILE, HISTORY_LIMIT),
		wsConn:   wsConn,
	}
}

func (c *CLI) Run(outputFile *os.File) error {
	if err := keyboard.Open(); err != nil {
		return err
	}
	defer keyboard.Close()
	defer c.history.SaveToFile()

	keysEvents, err := keyboard.GetKeys(KEYBOARD_BUFFER_SIZE)
	if err != nil {
		return err
	}

	fmt.Println("Connection Mode: Press ESC to enter Request mode")

	for {
		select {
		case event := <-keysEvents:
			//nolint:gomnd
			switch event.Key {
			case keyboard.KeyCtrlC, keyboard.KeyCtrlD:
				return nil

			case keyboard.KeyEsc:
				fmt.Println("Request Mode: Type your API request and press Ctrl+S to send it. Press ESC to cancel request")
				req, err := c.requestMode(keysEvents)

				if err != nil {
					if err.Error() == "interrupted" {
						return nil
					}

					fmt.Println(err)
				}

				if req != "" {
					err = c.wsConn.Send(req)
					if err != nil {
						fmt.Println("Fail to send request:", err)
					}
				}

				fmt.Println("Connection Mode: Press ESC to enter Request mode")
			default:
				continue
			}

		case msg := <-c.wsConn.Messages:

			output, err := c.formater.FormatMessage(msg)
			if err != nil {
				log.Printf("Fail to format message: %s, %s\n", err, msg.Data)
			}

			fmt.Printf("%s\n\n", output)

			if outputFile != nil {
				output, err := c.formater.FormatForFile(msg)
				if err != nil {
					log.Printf("Fail to format message for file: %s, %s\n", err, msg.Data)
				}
				fmt.Fprintln(outputFile, output)
			}
		}
	}
}

func (c *CLI) requestMode(keyStream <-chan keyboard.KeyEvent) (string, error) {
	buffer := ""

	historyIndex := 0

	for e := range keyStream {
		if e.Err != nil {
			return buffer, e.Err
		}

		//nolint:gomnd
		switch e.Key {
		case keyboard.KeyCtrlC, keyboard.KeyCtrlD:
			return buffer, fmt.Errorf("interrupted")
		case keyboard.KeyCtrlS:
			if buffer == "" {
				return buffer, fmt.Errorf("cannot send empty request")
			}
			requet := strings.TrimSpace(buffer)
			c.history.AddRequest(requet)
			return requet, nil
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

		case keyboard.KeyBackspace, keyboard.KeyDelete, MACOS_DELETE_KEY:
			if len(buffer) == 0 {
				continue
			}

			if buffer[len(buffer)-1] == '\n' {
				buffer = buffer[:len(buffer)-1]
				fmt.Print(LINE_UP)
				startPrevLine := strings.LastIndex(buffer, "\n")
				if startPrevLine == -1 {
					startPrevLine = 0
				} else {
					startPrevLine++
				}
				fmt.Print(buffer[startPrevLine:])
			} else {
				fmt.Print("\b \b")
				buffer = buffer[:len(buffer)-1]
			}
			continue
		case keyboard.KeyArrowUp:
			historyIndex++
			req := c.history.GetRequst(historyIndex)

			if req == "" {
				historyIndex--
				continue
			}

			c.clearInput(buffer)

			fmt.Print(req)
			buffer = req
			continue
		case keyboard.KeyArrowDown:
			historyIndex--
			req := c.history.GetRequst(historyIndex)

			if req == "" {
				historyIndex++
				continue
			}

			c.clearInput(buffer)

			fmt.Print(req)
			buffer = req
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

func (c *CLI) clearInput(buffer string) {
	for i := 0; i < len(buffer); i++ {
		if buffer[i] == '\n' {
			fmt.Print(LINE_UP)
			fmt.Print(LINE_CLEAR)
		} else {
			fmt.Print("\b \b")
		}
	}
}
