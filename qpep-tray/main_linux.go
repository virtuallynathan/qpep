//go:generate go build -o qpep-tray
package main

import (
	"github.com/getlantern/systray"
)

func main() {
	systray.Run(onReady, nil)
}
