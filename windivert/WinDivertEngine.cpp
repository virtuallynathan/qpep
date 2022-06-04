
extern "C" {
    #include "windivert_wrapper.h"
    #include "stdio.h"
    #include "stdarg.h"
    #include "stdlib.h"

    #include "engine.h"
}

connection connectionsList[65536];
SRWLOCK sharedRWLock;

HANDLE diverterHandle = INVALID_HANDLE_VALUE;
HANDLE threadHandles[MAX_THREADS];

int diveterMessagesEnabledToGo = TRUE;
const char* tempAddress = "192.168.1.10";

int InitializeWinDivertEngine(int port, int numThreads) 
{
    if( port < 1 || port > 65536 || numThreads < 1 || numThreads > MAX_THREADS ) {
        logNativeMessageToGo(0, "Cannot initialize windiver engine with provided data, port:%d, threads:%d", port, numThreads);
        return DIVERT_ERROR_FAILED;
    }

    logNativeMessageToGo(0, "Initializing windivert engine...");
    InitializeSRWLock(&sharedRWLock);

    char filterOut[256] = "";
    snprintf(filterOut, 256, FILTER_OUTBOUND, port, 443, 443);
    logNativeMessageToGo(0, "Filtering outbound with %s", filterOut);

    diverterHandle = WinDivertOpen( filterOut, WINDIVERT_LAYER_NETWORK, 0, 0 );
    if (diverterHandle == INVALID_HANDLE_VALUE) {
        logNativeMessageToGo(0, "OUT: Could not initialize windivert engine, errorcode %d", GetLastError());
        return DIVERT_ERROR_NOTINITILIZED;
    }

    for( int i=0; i<MAX_THREADS; i++ ) {
        threadHandles[i] = INVALID_HANDLE_VALUE;
    }
    for( int i=0; i<65536; i++ ) {
        connectionsList[i].origSrcPort = 0;
        connectionsList[i].origDstPort = 0;
        connectionsList[i].state = STATE_CLOSED;
    }

    for( int i=0; i<numThreads; i++ ) {
        threadParameters* th = (threadParameters*)malloc( sizeof(threadParameters) );
        th->gatewayAddress = (char*)tempAddress;
        th->gatewayPort = port;
        th->threadID = i;

        threadHandles[i] = CreateThread(
            NULL,    // Thread attributes
            0,       // Stack size (0 = use default)
            dispatchDivertedOutboundPackets, // Thread start address
            (void*)th,    // Parameter to pass to the thread
            0,       // Creation flags
            NULL );   // Thread id

        if (threadHandles[i] == NULL) { // Thread creation failed.
            logNativeMessageToGo(0, "Could not initialize windivert engine, errorcode %d", GetLastError());
            return DIVERT_ERROR_FAILED;
        }
    }

    return DIVERT_OK;
}

int CloseWinDivertEngine() 
{
    if( diverterHandle == INVALID_HANDLE_VALUE ) {
        logNativeMessageToGo(0, "Windivert engine must first be initialized");
        return DIVERT_ERROR_NOTINITILIZED;
    }

    logNativeMessageToGo(0, "Closing windivert engine...");

    // stops new incoming packets
    WinDivertShutdown(diverterHandle, WINDIVERT_SHUTDOWN_RECV);

    // threads will stop when the queue is empty
    int resultThread = TRUE;
    logNativeMessageToGo(0, "Waiting for the divert engine to stop...");
    for( int i=0; i<MAX_THREADS; i++ ) {
        if( threadHandles[i] == INVALID_HANDLE_VALUE )
            continue;

        WaitForSingleObject(threadHandles[i], INFINITE);
        int result = CloseHandle(threadHandles[i]);
        threadHandles[i] = INVALID_HANDLE_VALUE;

        if( result != TRUE )
            resultThread = result;
    }

    // close resources and check result
    int resultDivert= WinDivertClose(diverterHandle);
    diverterHandle = INVALID_HANDLE_VALUE;

    if( resultDivert!= TRUE || resultThread != TRUE ) {
        logNativeMessageToGo(0, "Could not stop the engine, errorcode: %d, status: %d/%d", 
            GetLastError(), resultDivert, resultThread);
        return DIVERT_ERROR_FAILED;
    }

    return DIVERT_OK;
}

DWORD WINAPI dispatchDivertedOutboundPackets(LPVOID lpParameter)
{
    WinDivertSetParam(diverterHandle, WINDIVERT_PARAM_QUEUE_LENGTH, 8192);
    WinDivertSetParam(diverterHandle, WINDIVERT_PARAM_QUEUE_TIME, 1024);
        
    WINDIVERT_ADDRESS recv_addr; // Packet address
    WINDIVERT_ADDRESS send_addr; // Packet address

    unsigned char* packet = (unsigned char*)malloc(sizeof(unsigned char) * (MAXBUF+1));    // Packet buffer
    UINT packetLen = 0;

    // parse structures
    PWINDIVERT_IPHDR ip_header;
    PWINDIVERT_IPV6HDR ipv6_header;
    PWINDIVERT_TCPHDR tcp_header;
    PVOID packetData;
    UINT packetDataLen = 0;
    PVOID next = NULL;
    UINT nextLen = 0;

    char src_str[INET6_ADDRSTRLEN+1], dst_str[INET6_ADDRSTRLEN+1];

    threadParameters* th = (threadParameters*)lpParameter;
    
    int error = 0;

    while (TRUE)
    {
        if (!WinDivertRecv(diverterHandle, (void*)packet, MAXBUF, &packetLen, &recv_addr))
        {
            error = GetLastError();
            if( error == ERROR_NO_DATA ) {
                logNativeMessageToGo(th->threadID,  "WinDivertRecv no more data");
                free( th );
                return 0;
            }

            logNativeMessageToGo(th->threadID, "WinDivertRecv returned error %d", error);
            continue;
        }

        logNativeMessageToGo( th->threadID, ">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>" );
        logNativeMessageToGo(th->threadID,  "WinDivertRecv: Received packet (%d)", packetLen );

        ip_header = NULL;
        ipv6_header = NULL;
        tcp_header = NULL;
        packetData = NULL;
        packetDataLen = 0;
        next = NULL;
        nextLen = 0;

        // common packet parse
        WinDivertHelperParsePacket(packet, packetLen, 
            &ip_header, &ipv6_header,
            NULL, NULL, NULL, 
            &tcp_header, 
            NULL, 
            &packetData, &packetDataLen,
            &next, &nextLen);

        // common headers check
        if (ip_header == NULL && ipv6_header == NULL)
        {
            logNativeMessageToGo(th->threadID,  "WinDivertRecv: Received non-ip packet" );
            continue;
        }
        if( ip_header != NULL )
        {
            WinDivertHelperFormatIPv4Address(ntohl(ip_header->SrcAddr),
                src_str, sizeof(src_str));
            WinDivertHelperFormatIPv4Address(ntohl(ip_header->DstAddr),
                dst_str, sizeof(dst_str));
        }
        if( ipv6_header != NULL )
        {
            UINT32 src_addr[4], dst_addr[4];

            WinDivertHelperNtohIPv6Address(ipv6_header->SrcAddr, src_addr);
            WinDivertHelperNtohIPv6Address(ipv6_header->DstAddr, dst_addr);
            WinDivertHelperFormatIPv6Address(src_addr, src_str,
                sizeof(src_str));
            WinDivertHelperFormatIPv6Address(dst_addr, dst_str,
                sizeof(dst_str));
        }

        if( next != NULL ) {
            // TODO
            logNativeMessageToGo(th->threadID,  "WinDivertRecv: next len (%d)", nextLen );
        }

        UINT localSrcPort = ntohs(tcp_header->SrcPort);
        UINT localDstPort = ntohs(tcp_header->DstPort);

        if( !recv_addr.Outbound ) {
            logNativeMessageToGo(th->threadID, "Dropped...");
            continue;
        }

        // announce and handle the packet
        logNativeMessageToGo(th->threadID,  "Received packet: %s:%d %s %s:%d (%d) [S:%d A:%d F:%d P:%d R:%d]", 
            src_str, ntohs(tcp_header->SrcPort), 
            (recv_addr.Outbound? "---->": "<----"),
            dst_str, ntohs(tcp_header->DstPort),
            ntohs(tcp_header->SeqNum),
            tcp_header->Syn, tcp_header->Ack,
            tcp_header->Fin, tcp_header->Psh,
            tcp_header->Rst );

        dumpPacket( (PVOID)packet, packetLen );

        memcpy(&send_addr, &recv_addr, sizeof(WINDIVERT_ADDRESS));

        if( localSrcPort != th->gatewayPort ) {
            logNativeMessageToGo(th->threadID, "LOCAL -> GO");
            // local to go listener redirect
            BOOL redirected = handleLocalToServerPacket( th, ip_header, ipv6_header, tcp_header, 
                &recv_addr, &send_addr );

            if( !redirected )
                continue;
        } else {
            logNativeMessageToGo(th->threadID, "GO -> LOCAL");
            // go listener to local redirect
            BOOL redirected = handleServerToLocalPacket( th, ip_header, ipv6_header, tcp_header, 
                &recv_addr, &send_addr );

            if( !redirected )
                continue;
        }
        send_addr.Impostor = 1;

        // Reparsing for updated data
        WinDivertHelperParsePacket(packet, packetLen, 
            &ip_header, &ipv6_header,
            NULL, NULL, NULL, 
            &tcp_header, 
            NULL, 
            &packetData, &packetDataLen,
            &next, &nextLen);

        // announce new packet
        logNativeMessageToGo(th->threadID,  "Redirected packet: %s:%d %s %s:%d (%d) [S:%d A:%d F:%d P:%d R:%d]", 
            src_str, ntohs(tcp_header->SrcPort), 
            (recv_addr.Outbound? "---->": "<----"),
            dst_str, ntohs(tcp_header->DstPort),
            ntohs(tcp_header->SeqNum),
            tcp_header->Syn, tcp_header->Ack,
            tcp_header->Fin, tcp_header->Psh,
            tcp_header->Rst );

        dumpPacket( (PVOID)packet, packetLen );

        if( WinDivertHelperCalcChecksums(packet, packetLen, &send_addr, 0) != TRUE ) 
        {
            logNativeMessageToGo(th->threadID, "Could not calculate checksum, dropping...");
            continue;
        }

        logNativeMessageToGo(th->threadID,  "WinDivertSend: Write packet (%d)", packetLen );
        if (!WinDivertSend(diverterHandle, packet, packetLen, NULL, &send_addr))
        {
            error = GetLastError();
            if( error == ERROR_NO_DATA ) {
                logNativeMessageToGo(th->threadID,  "WinDivertSend no more data" );
                return 0;
            }

            logNativeMessageToGo(th->threadID, "WinDivertSend returned error %d\n", error);
            continue;
        }

        logNativeMessageToGo(th->threadID,  "WinDivertSend: Sent packet" );
    }

    return 0;
}

// eg. 54731 -> 8080 redirected as 54731 -> 9443 (connection index 54731)
BOOL handleLocalToServerPacket(
    threadParameters* th,
    PWINDIVERT_IPHDR ip_header,
    PWINDIVERT_IPV6HDR ipv6_header,
    PWINDIVERT_TCPHDR tcp_header,
    WINDIVERT_ADDRESS* recv_addr,
    WINDIVERT_ADDRESS* send_addr ) 
{
    if( !acquireLock(FALSE) ) {
        logNativeMessageToGo(th->threadID,  "Lock acquire timeout, dropping packet...");
        return FALSE;
    }
    int portSrcIdx = (int)ntohs(tcp_header->SrcPort);
    int portDstIdx = (int)ntohs(tcp_header->DstPort);
    UINT connState = connectionsList[ portSrcIdx ].state;
    UINT connStatePrev = connState;
    releaseLock(FALSE);

    // TODO: handle FYN
    logNativeMessageToGo(th->threadID,  "Connection state for port %d: %s", portSrcIdx, connStateToString(connState));
    switch( connState ) {
        case STATE_CLOSED:
            if( !tcp_header->Syn ) {
                logNativeMessageToGo(th->threadID,  "Out-of-sequence packet for handshake, dropping...");
                return FALSE;
            }
            connState = STATE_SYN;
            break;

        case STATE_SYN:
            logNativeMessageToGo(th->threadID,  "Out-of-sequence packet for handshake, dropping...");
            return FALSE;

        case STATE_SYN_ACK:
            if( !tcp_header->Ack ) {
                logNativeMessageToGo(th->threadID,  "Out-of-sequence packet for handshake, dropping...");
                return FALSE;
            }
            connState = STATE_OPEN;
            break;

        case STATE_OPEN:
            logNativeMessageToGo(th->threadID,  "" );
            break;
    }

    BOOL isIPV4 = ( ip_header != NULL ) ? TRUE : FALSE;
    BOOL isIPV6 = ( ipv6_header != NULL ) ? TRUE : FALSE;

    if( connState != connStatePrev ) {
        logNativeMessageToGo(th->threadID,  "NEW Connection state for port %d: %s", portSrcIdx, connStateToString(connState));

        // update state and set ports if necessary
        if( !acquireLock(TRUE) ) {
            logNativeMessageToGo(th->threadID,  "Lock acquire timeout, dropping packet...");
            return FALSE;
        }
        connectionsList[ portSrcIdx ].state = connState;
        if( connectionsList[ portSrcIdx ].origSrcPort == 0 ) {
            connectionsList[ portSrcIdx ].origSrcPort = portSrcIdx;
            connectionsList[ portSrcIdx ].origDstPort = portDstIdx;
            if( isIPV4 ) {
                connectionsList[ portSrcIdx ].origSrcAddress = ip_header->SrcAddr;
                connectionsList[ portSrcIdx ].origDstAddress = ip_header->DstAddr;
            }
            if( isIPV6 ) {
                connectionsList[ portSrcIdx ].origSrcAddressV6[0] = ipv6_header->SrcAddr[0];
                connectionsList[ portSrcIdx ].origSrcAddressV6[1] = ipv6_header->SrcAddr[1];
                connectionsList[ portSrcIdx ].origSrcAddressV6[2] = ipv6_header->SrcAddr[2];
                connectionsList[ portSrcIdx ].origSrcAddressV6[3] = ipv6_header->SrcAddr[3];

                connectionsList[ portSrcIdx ].origDstAddressV6[0] = ipv6_header->DstAddr[0];
                connectionsList[ portSrcIdx ].origDstAddressV6[1] = ipv6_header->DstAddr[1];
                connectionsList[ portSrcIdx ].origDstAddressV6[2] = ipv6_header->DstAddr[2];
                connectionsList[ portSrcIdx ].origDstAddressV6[3] = ipv6_header->DstAddr[3];
            }
        }
        releaseLock(TRUE);
    }

    // redirect to the local go listener
    tcp_header->DstPort = htons(th->gatewayPort);

    if( isIPV4 )
    {
        UINT32 remote_addr;
        WinDivertHelperParseIPv4Address(th->gatewayAddress, &remote_addr);
        ip_header->DstAddr = htonl(remote_addr);
    }
    if( isIPV6 )
    {
        UINT32 remote_addr[4];
        WinDivertHelperParseIPv6Address(th->gatewayAddress, remote_addr);
        ipv6_header->DstAddr[0] = htonl(remote_addr[0]);
        ipv6_header->DstAddr[1] = htonl(remote_addr[1]);
        ipv6_header->DstAddr[2] = htonl(remote_addr[2]);
        ipv6_header->DstAddr[3] = htonl(remote_addr[3]);
    }

    // redirect data for device driver
    WinDivertHelperParseIPv4Address(th->gatewayAddress, send_addr->Flow.RemoteAddr);
    send_addr->Flow.RemotePort = th->gatewayPort;

    WinDivertHelperParseIPv4Address(th->gatewayAddress, send_addr->Socket.RemoteAddr);
    send_addr->Socket.RemotePort = th->gatewayPort;

    return TRUE;
}

// eg. 9443 -> 54731 redirected as 8080 -> 54731 (connection index 54731)
BOOL handleServerToLocalPacket(
    threadParameters* th,
    PWINDIVERT_IPHDR ip_header,
    PWINDIVERT_IPV6HDR ipv6_header,
    PWINDIVERT_TCPHDR tcp_header,
    WINDIVERT_ADDRESS* recv_addr,
    WINDIVERT_ADDRESS* send_addr ) 
{
    BOOL isIPV4 = ( ip_header != NULL ) ? TRUE : FALSE;
    BOOL isIPV6 = ( ipv6_header != NULL ) ? TRUE : FALSE;

    if( !acquireLock(FALSE) ) {
        logNativeMessageToGo(th->threadID,  "Lock acquire timeout, dropping packet...");
        return FALSE;
    }
    int portSrcIdx = (int)ntohs(tcp_header->DstPort);
    UINT connState = connectionsList[ portSrcIdx ].state;
    UINT connStatePrev = connState;
    UINT origDstPort = connectionsList[ portSrcIdx ].origDstPort;

    UINT32 origSrcAddressV4 = connectionsList[ portSrcIdx ].origSrcAddress;
    UINT32 origDstAddressV4 = connectionsList[ portSrcIdx ].origDstAddress;

    UINT32 origSrcAddressV6[4];
    origSrcAddressV6[0] = connectionsList[ portSrcIdx ].origSrcAddressV6[0];
    origSrcAddressV6[1] = connectionsList[ portSrcIdx ].origSrcAddressV6[1];
    origSrcAddressV6[2] = connectionsList[ portSrcIdx ].origSrcAddressV6[2];
    origSrcAddressV6[3] = connectionsList[ portSrcIdx ].origSrcAddressV6[3];

    UINT32 origDstAddressV6[4];
    origDstAddressV6[0] = connectionsList[ portSrcIdx ].origDstAddressV6[0];
    origDstAddressV6[1] = connectionsList[ portSrcIdx ].origDstAddressV6[1];
    origDstAddressV6[2] = connectionsList[ portSrcIdx ].origDstAddressV6[2];
    origDstAddressV6[3] = connectionsList[ portSrcIdx ].origDstAddressV6[3];
    releaseLock(FALSE);

    // TODO: handle FYN
    logNativeMessageToGo(th->threadID,  "Connection state for port %d: %s", portSrcIdx, connStateToString(connState));
    switch( connState ) {
        case STATE_CLOSED:
            logNativeMessageToGo(th->threadID,  "Connection can only be established from local side, dropped");
            return FALSE;

        case STATE_SYN:
            if( !(tcp_header->Syn && tcp_header->Ack) ) {
                logNativeMessageToGo(th->threadID,  "Out-of-sequence packet for handshake, dropping...");
                return FALSE;
            }
            connState = STATE_SYN_ACK;
            break;

        case STATE_SYN_ACK:
            if( !(tcp_header->Ack) ) {
                logNativeMessageToGo(th->threadID,  "Out-of-sequence packet for handshake, dropping...");
                return FALSE;
            }
            break;

        case STATE_OPEN:
            logNativeMessageToGo(th->threadID,  "" );
    }

    if( connState != connStatePrev ) {
        logNativeMessageToGo(th->threadID,  "NEW Connection state for port %d: %s", portSrcIdx, connStateToString(connState));

        // connection state update
        if( !acquireLock(TRUE) ) {
            logNativeMessageToGo(th->threadID,  "Lock acquire timeout, dropping packet...");
            return FALSE;
        }
        connectionsList[ portSrcIdx ].state = connState;
        releaseLock(TRUE);
    }

    // redirect from local go listener as the original source port
    tcp_header->SrcPort = htons(origDstPort);
    if( isIPV4 ) {
        ip_header->SrcAddr = origDstAddressV4;
        ip_header->DstAddr = origSrcAddressV4;

        ipV4PackedToUnpackedNetworkByteOrder( origDstAddressV4, send_addr->Flow.LocalAddr );
        ipV4PackedToUnpackedNetworkByteOrder( origSrcAddressV4, send_addr->Flow.RemoteAddr );
    }
    if( isIPV6 ) {
        ipv6_header->SrcAddr[0] = origDstAddressV6[0];
        ipv6_header->SrcAddr[1] = origDstAddressV6[1];
        ipv6_header->SrcAddr[2] = origDstAddressV6[2];
        ipv6_header->SrcAddr[3] = origDstAddressV6[3];

        ipv6_header->DstAddr[0] = origSrcAddressV6[0];
        ipv6_header->DstAddr[1] = origSrcAddressV6[1];
        ipv6_header->DstAddr[2] = origSrcAddressV6[2];
        ipv6_header->DstAddr[3] = origSrcAddressV6[3];

        send_addr->Flow.LocalAddr[0] = origDstAddressV6[0];
        send_addr->Flow.LocalAddr[1] = origDstAddressV6[1];
        send_addr->Flow.LocalAddr[2] = origDstAddressV6[2];
        send_addr->Flow.LocalAddr[3] = origDstAddressV6[3];

        send_addr->Flow.RemoteAddr[0] = origSrcAddressV6[0];
        send_addr->Flow.RemoteAddr[1] = origSrcAddressV6[1];
        send_addr->Flow.RemoteAddr[2] = origSrcAddressV6[2];
        send_addr->Flow.RemoteAddr[3] = origSrcAddressV6[3];
    }

    send_addr->Flow.LocalPort = htons(origDstPort);
    send_addr->Socket.LocalPort = htons(origDstPort);

    return TRUE;
}

// assumes network (big-endian) byte order
void ipV4PackedToUnpackedNetworkByteOrder( UINT32 packed, UINT32* unpacked ) {
    unpacked[0] = (UINT)(packed & (0xFF000000) >> 18);
    unpacked[1] = (UINT)(packed & (0x00FF0000) >> 10);
    unpacked[2] = (UINT)(packed & (0x0000FF00) >> 8);
    unpacked[3] = (UINT)(packed & (0x000000FF));
}

void dumpPacket( PVOID packetData, UINT len ) {
    printf("%06X ", 0);
    for( int i=0; i<len; i++ ) {
        if( i > 0 && i % 16 == 0 )
            printf("\n%06X ", i-16);
        printf("%02X ", ((unsigned char*)packetData)[i]);
    }
    printf("\n");
}

void  logNativeMessageToGo(int thid, const char* format...) {
    if( !diveterMessagesEnabledToGo )
        return;

    char buffer[8 + 1024] = "";
    snprintf(buffer, 1024, "[%04d] ", thid);

    va_list _ArgList;
    va_start(_ArgList, format);
    vsnprintf(buffer + 7, 1024, format, _ArgList);
    va_end(_ArgList);

    logMessageToGo( buffer );
}

void EnableMessageOutputToGo( int enabled ) {
    if( enabled > 0 )
        enabled = TRUE;
    if( enabled < 0 )
        enabled = FALSE;
    diveterMessagesEnabledToGo = enabled;
}

BOOLEAN acquireLock( BOOLEAN exclusive ) {
    BOOLEAN acquired = FALSE;
    for( int i=0; i<100; i++ ) {
        if( exclusive )
            acquired = TryAcquireSRWLockExclusive(&sharedRWLock);
        else
            acquired = TryAcquireSRWLockShared(&sharedRWLock);

        if( acquired != TRUE )
            Sleep( 1 );
        else
            break;
    }

    return acquired;
}

void releaseLock( BOOLEAN exclusive ) {
    if( exclusive ) {
        ReleaseSRWLockExclusive(&sharedRWLock);
        return;
    }

    ReleaseSRWLockShared(&sharedRWLock);
}

const char* connStateToString( UINT state ) {
    switch( state ) {
        case STATE_CLOSED:
            return "closed";
        case STATE_SYN:
            return "handshaking syn";
        case STATE_SYN_ACK:
            return "handshaking ack";
        case STATE_OPEN:
            return "open";
    }
    return "unknown";
}
