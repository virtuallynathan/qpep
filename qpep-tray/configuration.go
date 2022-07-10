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

	"github.com/skratchdot/open-golang/open"
	"gopkg.in/yaml.v3"
)

const (
	CONFIGFILENAME = "qpep-tray.yml"
	CONFIGPATH     = "qpep-tray"
	DEFAULTCONFIG  = `acks: 10
ackDelay: 25
congestion: 4
decimate: 4
minBeforeDecimation: 100
gateway: 198.18.0.254
port: 443
apiport: 444
listenaddress: 192.168.1.10
listenport: 9443
multistream: true
verbose: false
varAckDelay: 0
threads: 1
`
)

type QPepConfigYAML struct {
	Acks             int    `yaml:"acks"`
	AckDelay         int    `yaml:"ackDelay"`
	Congestion       int    `yaml:"congestion"`
	Decimate         int    `yaml:"decimate"`
	DelayDecimate    int    `yaml:"minBeforeDecimation"`
	GatewayHost      string `yaml:"gateway"`
	GatewayPort      int    `yaml:"port"`
	GatewayAPIPort   int    `yaml:"apiport"`
	ListenHost       string `yaml:"listenaddress"`
	ListenPort       int    `yaml:"listenport"`
	MultiStream      bool   `yaml:"multistream"`
	Verbose          bool   `yaml:"verbose"`
	VarAckDelay      int    `yaml:"varAckDelay"`
	WinDivertThreads int    `yaml:"threads"`
}

var qpepConfig QPepConfigYAML

func readConfiguration() (outerr error) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("PANIC: ", err)
			debug.PrintStack()
			outerr = errors.New(fmt.Sprintf("%v", err))
		}
	}()

	basedir := os.Getenv(BASEDIR_ENVIRONMENTVAR)
	confdir := filepath.Join(basedir, CONFIGPATH)
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
	if err := yaml.Unmarshal(data, &qpepConfig); err != nil {
		ErrorMsg("Could not decode configuration file: %v", err)
		return err
	}

	log.Println("Configuration Loaded")
	return nil
}

func getConfFile() string {
	basedir := os.Getenv(BASEDIR_ENVIRONMENTVAR)
	return filepath.Join(basedir, CONFIGPATH, CONFIGFILENAME)
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
				stat, err := os.Stat(confFile)
				if err != nil {
					continue
				}
				if !stat.ModTime().After(lastModTime) {
					continue
				}
				lastModTime = stat.ModTime()
				if ok := ConfirmMsg("Do you want to reload the configuration?"); !ok {
					continue
				}
				if readConfiguration() == nil {
					reloadClientIfRunning()
					reloadServerIfRunning()
				}
				continue
			}
		}
	}()

	return ctx, cancel
}
