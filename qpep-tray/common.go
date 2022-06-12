package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/getlantern/systray"
	"github.com/parvit/qpep/qpep-tray/icons"

	. "github.com/sqweek/dialog"
)

var (
	ErrFailed            = errors.New("failed")
	ErrNoCommand         = errors.New("could not create command")
	ErrCommandNotStarted = errors.New("could not start command")
)

func ErrorMsg(message string, parameters ...interface{}) {
	str := fmt.Sprintf(message, parameters...)
	log.Println("ERR: ", str)
	Message(str).Error()
}
func InfoMsg(message string, parameters ...interface{}) {
	str := fmt.Sprintf(message, parameters...)
	log.Println("INFO: ", str)
	Message(str).Info()
}
func ConfirmMsg(message string, parameters ...interface{}) bool {
	str := fmt.Sprintf(message, parameters...)
	log.Println("ASK: ", str)
	return Message(str).YesNo()
}

var contextWatchdog context.Context
var cancelWatchdog context.CancelFunc

func onReady() {
	contextWatchdog, cancelWatchdog = startReloadConfigurationWatchdog()

	systray.SetTemplateIcon(icons.MainIconData, icons.MainIconData)
	systray.SetTitle("QPep")
	systray.SetTooltip("TCP Accelerator")

	// We can manipulate the systray in other goroutines
	go func() {
		systray.SetTemplateIcon(icons.MainIconData, icons.MainIconData)
		systray.SetTitle("QPep TCP accelerator")
		systray.SetTooltip("QPep TCP accelerator")

		mConfig := systray.AddMenuItem("Edit Configuration", "Open configuration for next client / server executions")
		mConfigRefresh := systray.AddMenuItem("Reload Configuration", "Reload configuration from disk and restart the service")
		systray.AddSeparator()
		mClient := systray.AddMenuItemCheckbox("Client Disabled", "Launch/Stop QPep Client", false)
		mServer := systray.AddMenuItemCheckbox("Server Disabled", "Launch/Stop QPep Server", false)
		systray.AddSeparator()
		mQuit := systray.AddMenuItem("Quit", "Stop all and quit the whole app")

		// Sets the icon of a menu item.
		mQuit.SetIcon(icons.ExitIconData)
		mConfig.SetIcon(icons.ConfigIconData)
		mConfigRefresh.SetIcon(icons.RefreshIconData)

		mClientActive := false
		mServerActive := false

		for {
			select {
			case <-mConfig.ClickedCh:
				openConfigurationWithOSEditor()
				continue

			case <-mConfigRefresh.ClickedCh:
				readConfiguration()
				continue

			case <-mClient.ClickedCh:
				if !mClientActive {
					if ok := ConfirmMsg("Do you want to enable the client?"); !ok {
						break
					}
					if startClient() == nil {
						mClientActive = true
						mClient.SetTitle("Client Enabled")
						mClient.Enable()
						mClient.Check()

						mServerActive = false
						mServer.SetTitle("Server Disabled")
						mServer.Uncheck()
						mServer.Disable()
						stopServer()
					}

				} else {
					if ok := ConfirmMsg("Do you want to disable the client?"); !ok {
						break
					}
					if stopClient() == nil {
						mClientActive = false
						mClient.SetTitle("Client Disabled")
						mClient.Enable()
						mClient.Uncheck()

						mServerActive = false
						mServer.SetTitle("Server Disabled")
						mServer.Uncheck()
						mServer.Enable()
						stopServer()
					}
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
					startServer()

					mClientActive = false
					mClient.SetTitle("Client Disabled")
					mClient.Uncheck()
					mClient.Disable()
					stopClient()
				} else {
					if ok := ConfirmMsg("Do you want to enable the server?"); !ok {
						break
					}
					mServerActive = false
					mServer.SetTitle("Server Disabled")
					mServer.Enable()
					mServer.Uncheck()
					stopServer()

					mClientActive = false
					mClient.SetTitle("Client Disabled")
					mClient.Uncheck()
					mClient.Enable()
					stopClient()
				}

			case <-mQuit.ClickedCh:
				if ok := ConfirmMsg("Do you want to quit QPep and stop its services?"); !ok {
					break
				}

				stopClient()
				stopServer()

				systray.Quit()
				return
			}
		}
	}()
}

func onExit() {
	log.Println("Waiting for resources to be freed...")

	cancelWatchdog()
	select {
	case <-time.After(10 * time.Second):
		break
	case <-contextWatchdog.Done():
		break
	}

	log.Println("Closing...")
}
