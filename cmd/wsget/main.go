package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/ksysoev/wsget/pkg/cli"
	"github.com/ksysoev/wsget/pkg/ws"
)

var wsURL string
var insecure *bool
var OutputFH *os.File
var InputFH *os.File

func init() {
	url := flag.String("u", "", "Websocket url that will be used for connection. this argument is required")
	outputFile := flag.String("o", "", "Output file for saving requests and responses")
	showHelp := flag.Bool("h", false, "Prints this help message")
	insecure = flag.Bool("insecure", false, "Skip SSL certificate verification")

	flag.Parse()

	if (showHelp != nil && *showHelp) || (url == nil || *url == "") {
		flag.Usage()
		os.Exit(0)
	}

	wsURL = *url

	if outputFile != nil && *outputFile != "" {
		var err error

		OutputFH, err = os.Create(*outputFile)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func main() {
	wsInsp, err := ws.NewWS(wsURL, ws.Options{SkipSSLVerification: *insecure})
	if err != nil {
		log.Fatal(err)
	}

	defer wsInsp.Close()

	client := cli.NewCLI(wsInsp)

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

	err = client.Run(OutputFH)
	if err != nil {
		log.Println("Error:", err)
	}
}
