//go:generate go build -o qpep-tray
package main

import (
	"github.com/getlantern/systray"
	"log"
)

func readConfiguration() (outerr error) {
	return readConfigurationFromFile("HOME")
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
