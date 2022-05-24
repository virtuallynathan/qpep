package windivert

//#cgo windows CPPFLAGS: -D WIN32 -I include/
//#cgo windows,amd64 LDFLAGS: windivert/x64/WinDivert.dll
//#cgo windows,386 LDFLAGS: windivert/x86/WinDivert.dll
//#include "windivert_wrapper.h"
import "C"

func InitializeWinDivertEngine() int {
	return int(C.InitializeWinDivertEngine())
}
