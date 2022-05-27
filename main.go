package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime/debug"
	"time"

	"github.com/parvit/qpep/client"
	"github.com/parvit/qpep/server"
	"github.com/parvit/qpep/shared"

	"github.com/parvit/qpep/windivert"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("PANIC: %v", err)
			debug.PrintStack()
		}
	}()
	log.Println(windivert.InitializeWinDivertEngine("192.168.1.110", 9090, 4))

	client.ClientConfiguration.GatewayHost = shared.QuicConfiguration.GatewayIP

	execContext, cancelExecutionFunc := context.WithCancel(context.Background())

	if shared.QuicConfiguration.ClientFlag {
		fmt.Println("Running Client")
		go client.RunClient(execContext)
	} else {
		go server.RunServer(execContext)
	}

	interruptListener := make(chan os.Signal)
	signal.Notify(interruptListener, os.Interrupt)
	<-interruptListener

	cancelExecutionFunc()

	<-execContext.Done()

	log.Println("Shutdown...")
	log.Println(windivert.CloseWinDivertEngine())

	<-time.After(1 * time.Second)

	log.Println("Exiting...")
	os.Exit(1)
}
