
#ifndef WIN32
  #error Only supported compilation is on Windows platform
#endif // WIN32

// Including SDKDDKVer.h defines the highest available Windows platform.
// If you wish to build your application for a previous Windows platform, include WinSDKVer.h and
// set the _WIN32_WINNT macro to the platform you wish to support before including SDKDDKVer.h.
#include <SDKDDKVer.h>

#define WIN32_LEAN_AND_MEAN // Exclude rarely-used stuff from Windows headers
// Windows Header Files:
#include <windows.h>
#include <Synchapi.h>

#include "windivert.h"

enum {
  DIVERT_OK = 0,
  DIVERT_ERROR_NOTINITILIZED = 1,
  DIVERT_ERROR_ALREADY_INIT  = 2,
  DIVERT_ERROR_FAILED = 3,
};

extern int  InitializeWinDivertEngine(int port, int numThreads);
extern int  CloseWinDivertEngine();
extern void logMessageToGo( char* message );
extern void EnableMessageOutputToGo( int enabled );
