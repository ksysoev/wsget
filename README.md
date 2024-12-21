# wsget

[![Tests](https://github.com/ksysoev/wsget/actions/workflows/main.yml/badge.svg)](https://github.com/ksysoev/wsget/actions/workflows/main.yml)
[![codecov](https://codecov.io/gh/ksysoev/wsget/graph/badge.svg?token=JKPRCA5SSV)](https://codecov.io/gh/ksysoev/wsget)
[![Go Report Card](https://goreportcard.com/badge/github.com/ksysoev/wsget)](https://goreportcard.com/report/github.com/ksysoev/wsget)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

wsget is a command-line tool for interacting with a WebSocket server. It supports plain text and JSON messages and can save the output of the session into a file.

## Installation

### Downloading binaries:

Compiled executables can be downloaded from [here](https://github.com/ksysoev/wsget/releases).

### Install from source code:

```
go install github.com/ksysoev/wsget/cmd/wsget@latest
```

### Install with homebrew:

```
brew tap ksysoev/wsget
brew install wsget
```

## Usage

To use wsget, you need to specify the WebSocket URL:

```
wsget wss://ws.postman-echo.com/raw
```


You also can pass the initial request as part command line argument by using flag -r

```
wsget wss://ws.postman-echo.com/raw -r "Hello world!"
```


By default, wsget will print the data received from the WebSocket server only to the console. You can also save the data to a file using the -o flag:

```
wsget wss://ws.postman-echo.com/raw  -o output.txt
```

Example:

```
wsget "wss://ws.derivws.com/websockets/v3?app_id=1" -r '{"time":1}'
Use Enter to input request and send it, Ctrl+C to exit
->
{
  "time": 1
}
<-
{
  "echo_req": {
    "time": 1
  },
  "msg_type": "time",
  "time": 1698555261
}
```
## Editor Keyboard Shortcuts

### General Navigation and Editing Shortcuts

| Key/Combination | Action |
| --- |---|
| **Left Arrow** | Move the cursor one position to the left. |
| **Right Arrow** | Move the cursor one position to the right. |
| **Space** | Insert a space character at the current cursor position. |
| **Enter** | Adds a newline or completes editing (depends on editor mode). |
| **Backspace** | Remove the character before the cursor. |
| **Delete** | Remove the character after the cursor. |

### Advanced Navigation

| Key/Combination | Action |
|---| --- |
| **Alt + Backspace** | Delete the word to the left of the cursor. |
| **Alt + Left Arrow** | Move the cursor to the start of the previous word. |
| **Alt + Right Arrow** | Move the cursor to the start of the next word. |
| **Alt + Delete** | Delete the word to the right of the cursor. |
| **Home** | Move the cursor to the start of the line. |
| **End** | Move the cursor to the end of the line. |

### Content Modification

| Key/Combination | Action |
| --- | --- |
| **Ctrl + U** | Clear all content from the editor. |
| **Ctrl + L** | Clear the terminal's display while retaining content and positioning. |

### History Navigation

| Key/Combination | Action |
| --- | --- |
| **Up Arrow** | Cycle to the previous request in history. |
| **Down Arrow** | Cycle to the next request in history. |

### Miscellaneous Shortcuts

| Key/Combination | Action |
|---| --- |
| **Ctrl + S** | Complete editing. |
| **Ctrl + C** or **Ctrl + D** or **Esc** | Interrupt the editing process; cancel and terminate editing. |

## Macros

`wsget` provides a possibility for customization. You can create your sets of macros with a configuration file. the file should be located at `~/wsget/macro/your_configuration.yaml`. `wsget` will read all files from this directory and use only configuration files that match the WebSocket connection hostname.

```yaml
version: "1"
domains:
    - example.com
macro:
    ping:
        - |-
          send {
              "ping": 1,
          }
        - wait 5
```

### Primitive commands

- `edit {"ping": 1}` opens request editor with provided text
- `send {"ping": 1}` sends requests to WebSocket connection
- `wait 5` waits for responses or provided time out, whatever comes first. If the timeout is reached then an error will be returned. if `0` is provided command will wait response without a time limit
- `exit` interrupts the program execution
- `repeat 5 send {"ping": 1}` repeat provided command or macro defined number of times
- `sleep 1` sleeps for the provided number of seconds

### Macros arguments

Macro support [Go template language](https://pkg.go.dev/text/template). It provides a possibility to pass arguments to your macro command and substitute or adjust the behavior of your macro commands.

```
version: "1"
domains:
    - example.com
macro:
    authorize:
        - |-
          send {
              "authorize": "{{index .Args 0}}",
          }
        - wait 2
```

### Macros presets

- [Deriv API](https://github.com/ksysoev/wsget-deriv-api)

## License

wsget is licensed under the MIT License. See the LICENSE file for more information.