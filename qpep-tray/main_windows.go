//go:generate go build -ldflags -H=windowsgui -o qpep-tray.exe

package main

import (
	"github.com/getlantern/systray"
)

func main() {
	systray.Run(onReady, nil)
}
