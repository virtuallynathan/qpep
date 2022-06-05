//go:generate go build -ldflags -H=windowsgui -o qpep-tray.exe

package main

import (
	"log"
	"os"
	"runtime/debug"

	"github.com/getlantern/systray"
)

func readConfiguration() (outerr error) {
	return readConfigurationFromFile("APPDATA")
}

func main() {
	log.SetFlags(log.Ltime | log.Lmicroseconds)

	if err := readConfiguration(); err != nil {
		log.Println("Could not load configuration file: ", err)
		debug.PrintStack()
		os.Exit(1)
	}

	systray.Run(onReady, nil)

	log.Println("Closing...")
}
