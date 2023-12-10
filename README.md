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

### Macros presets

- [Deriv API](https://github.com/ksysoev/wsget-deriv-api)

## License

wsget is licensed under the MIT License. See the LICENSE file for more information.