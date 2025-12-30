package main

import (
	"flag"
	"log"

	ExampleClient "github.com/victorvbello/gomcp/example/client"
	ExampleServer "github.com/victorvbello/gomcp/example/server"
)

func main() {
	var startType string

	flag.StringVar(&startType, "t", "type", "Start type client/server")
	flag.Parse()

	switch startType {
	case "client":
		ExampleClient.ExampleEverythingWithSTDIOClient()
	case "server":
		ExampleServer.ExampleEverythingWithSTDIOServer()
	default:
		log.Printf("Start type %s not found", startType)
	}
}
