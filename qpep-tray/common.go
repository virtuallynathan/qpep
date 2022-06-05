package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"

	"github.com/getlantern/systray"
	"gopkg.in/yaml.v3"

	. "github.com/sqweek/dialog"
)

const (
	CONFIGFILENAME = "qpep-tray.yml"
	DEFAULTCONFIG  = `
ListenHost:        192.168.56.1
ListenPort:        9443
GatewayHost:       192.168.56.10
GatewayPort:       443
QuicStreamTimeout: 2
MultiStream:       true
IdleTimeout:       300s
ConnectionRetries: 3
WinDivertThreads:  1
Verbose:           false
`
)

type ClientConfigYAML struct {
	ListenHost        string
	ListenPort        int
	GatewayHost       string
	GatewayPort       int
	QuicStreamTimeout int
	MultiStream       bool
	IdleTimeout       string
	ConnectionRetries int
	WinDivertThreads  int
	Verbose           bool
}

var clientConfig ClientConfigYAML

func readConfigurationFromFile(baseDirEnvVar string) (outerr error) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("PANIC: ", err)
			debug.PrintStack()
			outerr = errors.New(fmt.Sprintf("%v", err))
		}
	}()

	basedir := os.Getenv(baseDirEnvVar)
	confdir := filepath.Join(basedir, ".qpep-tray")
	if _, err := os.Stat(confdir); errors.Is(err, os.ErrNotExist) {
		os.Mkdir(confdir, 0664)
	}

	confFile := filepath.Join(confdir, CONFIGFILENAME)
	if _, err := os.Stat(confFile); errors.Is(err, os.ErrNotExist) {
		os.WriteFile(confFile, []byte(DEFAULTCONFIG), 0664)
	}

	f, err := os.Open(confFile)
	if err != nil {
		panic(fmt.Sprintf("Could not read expected configuration file: %v", err))
	}
	defer func() {
		f.Close()
	}()

	data, err := io.ReadAll(f)
	if err != nil {
		panic(fmt.Sprintf("Could not read expected configuration file: %v", err))
	}
	if err := yaml.Unmarshal(data, &clientConfig); err != nil {
		panic(fmt.Sprintf("Could not decode expected configuration file: %v", err))
	}

	log.Println("Configuration Loaded")
	return nil
}

func onReady() {
	systray.SetTemplateIcon(MainIconData, MainIconData)
	systray.SetTitle("QPep")
	systray.SetTooltip("TCP Accelerator")

	// We can manipulate the systray in other goroutines
	go func() {
		systray.SetTemplateIcon(MainIconData, MainIconData)
		systray.SetTitle("QPep TCP accelerator")
		systray.SetTooltip("QPep TCP accelerator")

		mConfig := systray.AddMenuItem("Configuration", "Open configuration for next client / server executions")
		systray.AddSeparator()
		mClient := systray.AddMenuItemCheckbox("Client Disabled", "Launch/Stop QPep Client", false)
		mServer := systray.AddMenuItemCheckbox("Server Disabled", "Launch/Stop QPep Server", false)
		systray.AddSeparator()
		mQuit := systray.AddMenuItem("Quit", "Stop all and quit the whole app")

		// Sets the icon of a menu item.
		mQuit.SetIcon(ExitIconData)
		mConfig.SetIcon(ConfigIconData)

		mClientActive := false
		mServerActive := false

		for {
			select {
			case <-mConfig.ClickedCh:
				Message("Clicked").Info()
				fmt.Println("TODO") // TODO: Open client/server configuration file

			case <-mClient.ClickedCh:
				if !mClientActive {
					if ok := Message("Do you want to enable the client?").YesNo(); !ok {
						break
					}
					mClientActive = true
					mClient.SetTitle("Client Enabled")
					mClient.Enable()
					mClient.Check()

					mServerActive = false
					mServer.SetTitle("Server Disabled")
					mServer.Uncheck()
					mServer.Disable()
				} else {
					if ok := Message("Do you want to disable the client?").YesNo(); !ok {
						break
					}
					mClientActive = false
					mClient.SetTitle("Client Disabled")
					mClient.Enable()
					mClient.Uncheck()

					mServerActive = false
					mServer.SetTitle("Server Disabled")
					mServer.Uncheck()
					mServer.Enable()
				}

			case <-mServer.ClickedCh:
				if !mServerActive {
					if ok := Message("Do you want to enable the server?").YesNo(); !ok {
						break
					}
					mServerActive = true
					mServer.SetTitle("Server Enabled")
					mServer.Enable()
					mServer.Check()

					mClientActive = false
					mClient.SetTitle("Client Disabled")
					mClient.Uncheck()
					mClient.Disable()
				} else {
					if ok := Message("Do you want to enable the server?").YesNo(); !ok {
						break
					}
					mServerActive = false
					mServer.SetTitle("Server Disabled")
					mServer.Enable()
					mServer.Uncheck()

					mClientActive = false
					mClient.SetTitle("Client Disabled")
					mClient.Uncheck()
					mClient.Enable()
				}

			case <-mQuit.ClickedCh:
				if ok := Message("Do you want to quit QPep and stop its services?").YesNo(); !ok {
					break
				}
				systray.Quit()
				return
			}
		}
	}()
}

func onExit() {

}
