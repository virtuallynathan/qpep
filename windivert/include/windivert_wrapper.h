
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
  DIVERT_OK = 0,  //!< No issue
  DIVERT_ERROR_NOTINITILIZED = 1,  //!< InitializeWinDivertEngine was not called previously 
  DIVERT_ERROR_ALREADY_INIT  = 2,  //!< InitializeWinDivertEngine called again before CloseWinDivertEngine
  DIVERT_ERROR_FAILED = 3,         //!< Operation failed
  DIVERT_ERROR_NOT_OPEN = 4,       //!< Connection is not open so no state available
};

extern int  InitializeWinDivertEngine(char* host, int port, int numThreads);
extern int  CloseWinDivertEngine();
extern void logMessageToGo( char* message );
extern void EnableMessageOutputToGo( int enabled );
extern int  GetConnectionData( UINT sourcePort, UINT* origSrcPort, UINT* origDstPort, 
                               char* origSrcAddress, char* origDstAddress );
