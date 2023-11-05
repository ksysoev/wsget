# wsget

[![Tests](https://github.com/ksysoev/wsget/actions/workflows/main.yml/badge.svg)](https://github.com/ksysoev/wsget/actions/workflows/main.yml)
[![codecov](https://codecov.io/gh/ksysoev/wsget/graph/badge.svg?token=JKPRCA5SSV)](https://codecov.io/gh/ksysoev/wsget)
[![Go Report Card](https://goreportcard.com/badge/github.com/ksysoev/wsget)](https://goreportcard.com/report/github.com/ksysoev/wsget)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

wsget is a command-line tool for interacting with a WebSocket server. It supports plain text and  json messages, and can save the output of session into file.

## Installation

### Downloading binaries:

Compilied executables can be downloaded from [here](https://github.com/ksysoev/wsget/releases).

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


You also can pass initial request as part command line argument by using flag -r

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

## License

wsget is licensed under the MIT License. See the LICENSE file for more information.