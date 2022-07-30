package main

import (
	"log"
	"os/exec"

	"github.com/parvit/qpep/shared"
)

var clientCmd *exec.Cmd

func startClient() error {
	if clientCmd != nil {
		log.Println("ERROR: Cannot start an already running client, first stop it")
		return shared.ErrFailed
	}

	clientCmd = getClientCommand()

	if err := clientCmd.Start(); err != nil {
		ErrorMsg("Could not start client program: %v", err)
		clientCmd = nil
		return shared.ErrCommandNotStarted
	}
	InfoMsg("Client started")

	return nil
}

func stopClient() error {
	if clientCmd == nil {
		log.Println("ERROR: Cannot stop an already stopped client, first start it")
		return nil
	}

	if err := stopClientProcess(); err != nil {
		if ok := ConfirmMsg("Could not stop process gracefully (%v), do you want to force-terminate it?", err); !ok {
			return err
		}
		if err := clientCmd.Process.Kill(); err != nil {
			ErrorMsg("Could not force-terminate process")
			return err
		}
	}

	clientCmd.Wait()
	clientCmd = nil
	InfoMsg("Client stopped")
	return nil
}

func reloadClientIfRunning() {
	if clientCmd == nil {
		return
	}

	stopClient()
	startClient()
}
