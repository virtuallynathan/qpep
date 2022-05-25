package windivert

//#cgo windows CPPFLAGS: -D WIN32 -I include/
//#cgo windows,amd64 LDFLAGS: windivert/x64/WinDivert.dll
//#cgo windows,386 LDFLAGS: windivert/x86/WinDivert.dll
//#include "windivert_wrapper.h"
import "C"

const (
	DIVERT_OK                  = 0
	DIVERT_ERROR_NOTINITILIZED = 1
	DIVERT_ERROR_ALREADY_INIT  = 2
	DIVERT_ERROR_FAILED        = 3
)

func InitializeWinDivertEngine() int {
	return int(C.InitializeWinDivertEngine())
}

func CloseWinDivertEngine() int {
	return int(C.CloseWinDivertEngine())
}
