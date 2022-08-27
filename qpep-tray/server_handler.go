package main

import (
	"log"
	"os/exec"

	"github.com/parvit/qpep/shared"
)

var serverCmd *exec.Cmd

func startServer() error {
	if serverCmd != nil {
		log.Println("ERROR: Cannot start an already running server, first stop it")
		return shared.ErrFailed
	}

	serverCmd = getServerCommand()

	if err := serverCmd.Start(); err != nil {
		ErrorMsg("Could not start server program: %v", err)
		serverCmd = nil
		return shared.ErrCommandNotStarted
	}
	InfoMsg("Server started")

	return nil
}

func stopServer() error {
	if serverCmd == nil {
		log.Println("ERROR: Cannot stop an already stopped server, first start it")
		return nil
	}

	if err := stopServerProcess(); err != nil {
		log.Printf("Could not stop process gracefully (%v), will try to force-terminate it\n", err)

		if err := serverCmd.Process.Kill(); err != nil {
			ErrorMsg("Could not force-terminate process")
			return err
		}
	}

	serverCmd.Wait()
	serverCmd = nil
	InfoMsg("Server stopped")
	return nil
}

func reloadServerIfRunning() {
	if serverCmd == nil {
		return
	}

	stopServer()
	startServer()
}
