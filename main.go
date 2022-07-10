package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/parvit/qpep/api"
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

	execContext, cancelExecutionFunc := context.WithCancel(context.Background())

	if shared.QuicConfiguration.ClientFlag {
		runAsClient(execContext)
	} else {
		runAsServer(execContext)
	}

	interruptListener := make(chan os.Signal)
	signal.Notify(interruptListener, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-interruptListener

	cancelExecutionFunc()

	<-execContext.Done()

	log.Println("Shutdown...")
	log.Println(windivert.CloseWinDivertEngine())

	<-time.After(1 * time.Second)

	log.Println("Exiting...")
	os.Exit(1)
}

func runAsClient(execContext context.Context) {
	log.Println("Running Client")
	windivert.EnableDiverterLogging(client.ClientConfiguration.Verbose)

	gatewayHost := shared.QuicConfiguration.GatewayIP
	gatewayPort := shared.QuicConfiguration.GatewayPort
	listenHost := shared.QuicConfiguration.ListenIP
	listenPort := shared.QuicConfiguration.ListenPort
	threads := shared.QuicConfiguration.WinDivertThreads

	if code := windivert.InitializeWinDivertEngine(gatewayHost, listenHost, gatewayPort, listenPort, threads); code != windivert.DIVERT_OK {
		windivert.CloseWinDivertEngine()

		log.Printf("ERROR: Could not initialize WinDivert engine, code %d\n", code)
		os.Exit(1)
	}

	go client.RunClient(execContext)
}

func runAsServer(execContext context.Context) {
	log.Println("Running Server")
	go server.RunServer(execContext)
	go api.RunAPIServer(execContext)
}
