package main

import (
	"fmt"

	"github.com/getlantern/systray"
	"github.com/getlantern/systray/example/icon"
)

func onReady() {
	systray.SetTemplateIcon(icon.Data, icon.Data)
	systray.SetTitle("QPep")
	systray.SetTooltip("Lantern")

	// We can manipulate the systray in other goroutines
	go func() {
		systray.SetTemplateIcon(icon.Data, icon.Data)
		systray.SetTitle("QPep network accelerator")
		systray.SetTooltip("QPep network accelerator")

		mConfig := systray.AddMenuItem("Configuration", "Open configuration for next client / server executions")
		systray.AddSeparator()
		mDivert := systray.AddMenuItemCheckbox("Divert all", "Divert all outgoing traffic", true)
		mClient := systray.AddMenuItemCheckbox("Client Disabled", "Launch/Stop QPep Client", false)
		mServer := systray.AddMenuItemCheckbox("Server Disabled", "Launch/Stop QPep Server", false)
		systray.AddSeparator()
		mQuit := systray.AddMenuItem("Quit", "Stop all and quit the whole app")

		// Sets the icon of a menu item. Only available on Mac.
		mQuit.SetIcon(icon.Data)

		mDivertActive := true
		mClientActive := false
		mServerActive := false

		for {
			select {
			case <-mConfig.ClickedCh:
				fmt.Println("TODO") // TODO: Open client/server configuration file

			case <-mConfig.ClickedCh:
				if mDivertActive {
					mDivertActive = false
					mDivert.Uncheck()
				} else {
					mDivertActive = true
					mDivert.Check()
				}

			case <-mClient.ClickedCh:
				if !mClientActive {
					mClientActive = true
					mClient.SetTitle("Client Enabled")
					mClient.Enable()
					mClient.Check()

					mServerActive = false
					mServer.SetTitle("Server Disabled")
					mServer.Uncheck()
					mServer.Disable()
				} else {
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
					mServerActive = true
					mServer.SetTitle("Server Enabled")
					mServer.Enable()
					mServer.Check()

					mClientActive = false
					mClient.SetTitle("Client Disabled")
					mClient.Uncheck()
					mClient.Disable()
				} else {
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
				systray.Quit()
				return
			}
		}
	}()
}
