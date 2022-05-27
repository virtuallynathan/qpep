package windivert

//#cgo windows CPPFLAGS: -D WIN32 -I include/
//#cgo windows,amd64 LDFLAGS: windivert/x64/WinDivert.dll
//#cgo windows,386 LDFLAGS: windivert/x86/WinDivert.dll
//#include "windivert_wrapper.h"
import "C"

import (
	"log"
)

const (
	DIVERT_OK                  = 0
	DIVERT_ERROR_NOTINITILIZED = 1
	DIVERT_ERROR_ALREADY_INIT  = 2
	DIVERT_ERROR_FAILED        = 3
)

func InitializeWinDivertEngine(address string, port, numThreads int) int {
	return int(C.InitializeWinDivertEngine(C.CString(address), C.int(port), C.int(numThreads)))
}

func CloseWinDivertEngine() int {
	return int(C.CloseWinDivertEngine())
}

//export logMessageToGo
func logMessageToGo(msg *C.char) {
	log.Println(C.GoString(msg))
}
