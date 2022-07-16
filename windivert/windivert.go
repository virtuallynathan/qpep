//go:build !cgo
// +build !cgo

package windivert

import (
	"log"
)

const (
	DIVERT_OK                  = 0
	DIVERT_ERROR_NOTINITILIZED = 1
	DIVERT_ERROR_ALREADY_INIT  = 2
	DIVERT_ERROR_FAILED        = 3
)

func InitializeWinDivertEngine(gatewayAddr, listenAddr string, gatewayPort, listenPort, numThreads int) int {
	log.Println("WARNING: windivert package compiled without CGO") // message to check for failing CGO
	return DIVERT_OK
}

func CloseWinDivertEngine() int {
	log.Println("WARNING: windivert package compiled without CGO") // message to check for failing CGO
	return DIVERT_OK
}

func GetConnectionStateData(port int) (int, int, int, string, string) {
	log.Println("WARNING: windivert package compiled without CGO") // message to check for failing CGO
	return DIVERT_OK, -1, -1, "", ""
}

func EnableDiverterLogging(enable bool) {
	log.Println("WARNING: windivert package compiled without CGO") // message to check for failing CGO
	return
}
