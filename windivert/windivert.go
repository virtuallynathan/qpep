package windivert

//#cgo windows CPPFLAGS: -DWIN32 -D_WIN32_WINNT=0x0600 -I include/
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

func InitializeWinDivertEngine(port, numThreads int) int {
	return int(C.InitializeWinDivertEngine(C.int(port), C.int(numThreads)))
}

func CloseWinDivertEngine() int {
	return int(C.CloseWinDivertEngine())
}

func EnableDiverterLogging(enable bool) {
	if enable {
		log.Println("Diverter messages will be output")
		C.EnableMessageOutputToGo(C.int(1))
	} else {
		log.Println("Diverter messages will be ignored")
		C.EnableMessageOutputToGo(C.int(0))
	}
}

//export logMessageToGo
func logMessageToGo(msg *C.char) {
	log.Println(C.GoString(msg))
}
