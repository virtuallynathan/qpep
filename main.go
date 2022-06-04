package main

import (
	"context"
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

	log.SetFlags(log.Ltime | log.Lmicroseconds)

	client.ClientConfiguration.GatewayHost = shared.QuicConfiguration.GatewayIP
	client.ClientConfiguration.GatewayPort = shared.QuicConfiguration.GatewayPort
	client.ClientConfiguration.WinDivertThreads = shared.QuicConfiguration.WinDivertThreads
	client.ClientConfiguration.Verbose = shared.QuicConfiguration.Verbose

	execContext, cancelExecutionFunc := context.WithCancel(context.Background())

	if shared.QuicConfiguration.ClientFlag {
		log.Println("Running Client")
		windivert.EnableDiverterLogging(client.ClientConfiguration.Verbose)

		host := client.ClientConfiguration.GatewayHost
		port := client.ClientConfiguration.GatewayPort
		threads := client.ClientConfiguration.WinDivertThreads

		if windivert.InitializeWinDivertEngine(host, port, threads) != windivert.DIVERT_OK {
			windivert.CloseWinDivertEngine()
			os.Exit(1)
		}
		go client.RunClient(execContext)
	} else {
		log.Println("Running Server")
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
