package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/parvit/qpep/client"
	"github.com/parvit/qpep/server"
	"github.com/parvit/qpep/shared"

	"github.com/parvit/qpep/windivert"
)

func main() {
	log.Println(windivert.InitializeWinDivertEngine())

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
