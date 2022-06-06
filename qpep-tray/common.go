package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"

	"github.com/getlantern/systray"
	"github.com/skratchdot/open-golang/open"
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

func ErrorMsg(message string, parameters ...interface{}) {
	Message(fmt.Sprintf(message, parameters...)).Error()
}
func InfoMsg(message string, parameters ...interface{}) {
	Message(fmt.Sprintf(message, parameters...)).Info()
}
func ConfirmMsg(message string, parameters ...interface{}) bool {
	return Message(fmt.Sprintf(message, parameters...)).YesNo()
}

func readConfiguration() (outerr error) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("PANIC: ", err)
			debug.PrintStack()
			outerr = errors.New(fmt.Sprintf("%v", err))
		}
	}()

	basedir := os.Getenv(BASEDIR_ENVIRONMENTVAR)
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
		ErrorMsg("Could not read expected configuration file: %v", err)
		return err
	}
	defer func() {
		f.Close()
	}()

	data, err := io.ReadAll(f)
	if err != nil {
		ErrorMsg("Could not read expected configuration file: %v", err)
		return err
	}
	if err := yaml.Unmarshal(data, &clientConfig); err != nil {
		ErrorMsg("Could not decode configuration file: %v", err)
		return err
	}

	log.Println("Configuration Loaded")
	return nil
}

func getConfFile() string {
	basedir := os.Getenv(BASEDIR_ENVIRONMENTVAR)
	return filepath.Join(basedir, ".qpep-tray", CONFIGFILENAME)
}

func openConfigurationWithOSEditor() {
	confdir := getConfFile()

	if err := open.Run(confdir); err != nil {
		ErrorMsg("Editor configuration failed with error: %v", err)
		return
	}
}

func startReloadConfigurationWatchdog() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		confFile := getConfFile()

		var lastModTime time.Time
		if stat, err := os.Stat(confFile); err == nil {
			lastModTime = stat.ModTime()

		} else {
			ErrorMsg("Configuration file not found, stopping")
			cancel()
			return
		}

	CHECKLOOP:
		for {
			select {
			case <-ctx.Done():
				log.Println("Stopping configfile watchdog")
				break CHECKLOOP

			case <-time.After(1 * time.Second):
				if stat, err := os.Stat(confFile); err == nil {
					if !stat.ModTime().After(lastModTime) {
						continue
					}
					lastModTime = stat.ModTime()
					if ok := ConfirmMsg("Do you want to reload the configuration?"); ok {
						readConfiguration()
					}
				}
				continue
			}
		}
	}()

	return ctx, cancel
}

func onReady() {
	contextWatchdog, cancelWatchdog := startReloadConfigurationWatchdog()

	systray.SetTemplateIcon(MainIconData, MainIconData)
	systray.SetTitle("QPep")
	systray.SetTooltip("TCP Accelerator")

	// We can manipulate the systray in other goroutines
	go func() {
		systray.SetTemplateIcon(MainIconData, MainIconData)
		systray.SetTitle("QPep TCP accelerator")
		systray.SetTooltip("QPep TCP accelerator")

		mConfig := systray.AddMenuItem("Edit Configuration", "Open configuration for next client / server executions")
		mConfigReload := systray.AddMenuItem("Reload Configuration", "Reload configuration from disk and restart the service")
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
				openConfigurationWithOSEditor()
				continue

			case <-mConfigReload.ClickedCh:
				readConfiguration()
				continue

			case <-mClient.ClickedCh:
				if !mClientActive {
					if ok := ConfirmMsg("Do you want to enable the client?"); !ok {
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
					if ok := ConfirmMsg("Do you want to disable the client?"); !ok {
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
					if ok := ConfirmMsg("Do you want to enable the server?"); !ok {
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
					if ok := ConfirmMsg("Do you want to enable the server?"); !ok {
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
				if ok := ConfirmMsg("Do you want to quit QPep and stop its services?"); !ok {
					break
				}

				cancelWatchdog()
				select {
				case <-time.After(10 * time.Second):
					break
				case <-contextWatchdog.Done():
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
