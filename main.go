package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/virtuallynathan/qpep/client"
	"github.com/virtuallynathan/qpep/server"
	"github.com/virtuallynathan/qpep/shared"
)

func main() {
	client.ClientConfiguration.GatewayHost = shared.QuicConfiguration.GatewayIP

	if shared.QuicConfiguration.ClientFlag {
		fmt.Println("Running Client")
		go client.RunClient()
	} else {
		go server.RunServer()
	}

	interruptListener := make(chan os.Signal)
	signal.Notify(interruptListener, os.Interrupt)
	<-interruptListener

	log.Println("Exiting...")
	os.Exit(1)
}
