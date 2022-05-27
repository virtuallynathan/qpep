
extern "C" {
    #include "windivert_wrapper.h"
    #include "stdio.h"
    #include "stdarg.h"
}

#define MAXBUF 8192

HANDLE diverterHandle = INVALID_HANDLE_VALUE;
HANDLE threadHandles[16];

DWORD WINAPI HandleDivertedPackets(LPVOID lpParameter);
void  _LogMessageToGo(const char* format...);

int InitializeWinDivertEngine(char* address, int port, int numThreads) 
{
    if( diverterHandle != INVALID_HANDLE_VALUE ) {
        _LogMessageToGo("Windiver engine already initialized");
        return DIVERT_ERROR_ALREADY_INIT;
    }

    if( address == NULL || port > 65536 || numThreads < 1 || numThreads > 16 ) {
        _LogMessageToGo("Cannot initialize windiver engine with provided data");
        return DIVERT_ERROR_FAILED;
    }

    _LogMessageToGo("Initializing windivert engine...");

    char filter[256] = "";
    snprintf(filter, 256, FILTERFMT, "192.168.1.100", 9090);

    _LogMessageToGo(filter);

    diverterHandle = WinDivertOpen( filter, WINDIVERT_LAYER_NETWORK, WINDIVERT_PRIORITY_HIGHEST, 0 );
    if (diverterHandle == INVALID_HANDLE_VALUE) {
        _LogMessageToGo("Could not initialize windivert engine, errorcode %d", GetLastError());
        return DIVERT_ERROR_NOTINITILIZED;
    }

    for( int i=0; i<16; i++ ) {
        threadHandles[i] = INVALID_HANDLE_VALUE;
    }

    for( int i=0; i<numThreads; i++ ) {
        int index = i;
        threadHandles[i] = CreateThread(
            NULL,    // Thread attributes
            0,       // Stack size (0 = use default)
            HandleDivertedPackets, // Thread start address
            (void*)&index,    // Parameter to pass to the thread
            0,       // Creation flags
            NULL );   // Thread id

        if (threadHandles[i] == NULL) { // Thread creation failed.
            _LogMessageToGo("Could not initialize windivert engine, errorcode %d", GetLastError());
            for( int j=0; j<i; j++ ) {
                CloseHandle( threadHandles[j] );
                threadHandles[i] = INVALID_HANDLE_VALUE;
            }
            WinDivertClose(diverterHandle);
            diverterHandle = INVALID_HANDLE_VALUE;
            return DIVERT_ERROR_FAILED;
        }
    }

    return DIVERT_OK;
}

int CloseWinDivertEngine() 
{
    if( diverterHandle == INVALID_HANDLE_VALUE ) {
        _LogMessageToGo("Windivert engine must first be initialized");
        return DIVERT_ERROR_NOTINITILIZED;
    }

    _LogMessageToGo("Closing windivert engine...");

    // stops new incoming packets
    WinDivertShutdown(diverterHandle, WINDIVERT_SHUTDOWN_RECV);

    // threads will stop when the queue is empty
    int resultThread = TRUE;
    _LogMessageToGo("Waiting for the divert engine to stop...");
    for( int i=0; i<16; i++ ) {
        if( threadHandles[i] == INVALID_HANDLE_VALUE )
            continue;

        WaitForSingleObject(threadHandles[i], INFINITE);
        int result = CloseHandle(threadHandles[i]);
        threadHandles[i] = INVALID_HANDLE_VALUE;

        if( result != TRUE )
            resultThread = result;
    }

    // close resources and check result
    int resultDivert = WinDivertClose(diverterHandle);
    diverterHandle = INVALID_HANDLE_VALUE;

    if( resultDivert != TRUE || resultThread != TRUE ) {
        _LogMessageToGo("Could not stop the engine, errorcode: %d", GetLastError());
        return DIVERT_ERROR_FAILED;
    }

    return DIVERT_OK;
}

DWORD WINAPI HandleDivertedPackets(LPVOID lpParameter)
{
    WINDIVERT_ADDRESS addr; // Packet address
    char packet[MAXBUF];    // Packet buffer
    UINT packetLen = 0;

    int error = 0;

    while (TRUE)
    {
        if (!WinDivertRecv(diverterHandle, packet, sizeof(packet), &packetLen, &addr))
        {
            error = GetLastError();
            if( error == ERROR_NO_DATA ) {
                _LogMessageToGo( "WinDivertRecv no more data", NULL);
                return 0;
            }

            _LogMessageToGo("WinDivertRecv returned error %d\n", error, NULL);
            continue;
        }

        if (!WinDivertSend(diverterHandle, packet, packetLen, NULL, &addr))
        {
            error = GetLastError();
            if( error == ERROR_NO_DATA ) {
                printf( "WinDivertRecv no more data\n" );
                return 0;
            }

            _LogMessageToGo("WinDivertRecv returned error %d\n", error);
            continue;
        }
    }

    return 0;
}

void  _LogMessageToGo(const char* format...) 
{
    char buffer[8192] = "";

    va_list _ArgList;
    va_start(_ArgList, format);
    vsnprintf(buffer, 8192, format, _ArgList);
    va_end(_ArgList);

    logMessageToGo( buffer );
}
