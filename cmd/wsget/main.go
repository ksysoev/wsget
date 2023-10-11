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

var wsUrl string
var OutputFH *os.File
var InputFH *os.File

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

func main() {
	fmt.Println("Connecting to", wsUrl, "...")
	wsInsp, err := ws.NewWS(wsUrl)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected")
	defer wsInsp.Close()

	cli := cli.NewCLI(wsInsp)

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

	err = cli.Run(OutputFH)
	if err != nil {
		log.Fatal(err)
	}
}
