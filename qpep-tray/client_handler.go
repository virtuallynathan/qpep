package main

import (
	"log"
	"os/exec"
)

var clientCmd *exec.Cmd

func startClient() error {
	if clientCmd != nil {
		log.Println("ERROR: Cannot start an already running client, first stop it")
		return ErrFailed
	}

	clientCmd = getClientCommand()

	if err := clientCmd.Start(); err != nil {
		ErrorMsg("Could not start client program: %v", err)
		clientCmd = nil
		return ErrCommandNotStarted
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
		log.Printf("Could not stop process gracefully (%v), will try to force-terminate it\n", err)

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
