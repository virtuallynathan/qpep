
// definition for local dev
// #ifndef WIN32
// #define WIN32 
// #endif

#ifdef WIN32

extern "C" {
    #include "windivert_wrapper.h"
    #include "stdio.h"
    #include "stdarg.h"
    #include "stdlib.h"

    #include "engine.h"
}

connection connectionsList[65536]; //!< List of connection tracking for every source port
SRWLOCK sharedRWLock; //!< Synchronization lock for the worker threads, very similar to go's sync.RWMutex

HANDLE diverterHandle = INVALID_HANDLE_VALUE; //!< WinDivert handler
HANDLE threadHandles[MAX_THREADS]; //!< Thread handles

int diveterMessagesEnabledToGo = TRUE; //!< When true, verbose redirect messages are output in the go log

/**
 * @brief Initializes the divert engine and the worker threads to handle the packets
 * 
 * @param gatewayHost   Host of the remote qpep server
 * @param listenHost    Address on which the qpep client is listening
 * @param gatewayPort   Port of the remote qpep server
 * @param listenPort    Port of the local listening client
 * @param numThreads    Number of worker threads to use (1-8)
 * @return DIVERT_OK    if everything ok, an error otherwise
 */
int InitializeWinDivertEngine(char* gatewayHost, char* listenHost, int gatewayPort, int listenPort, int numThreads) 
{
    if( gatewayPort < 1 || gatewayPort > 65536 || numThreads < 1 || numThreads > MAX_THREADS ) {
        logNativeMessageToGo(0, "Cannot initialize windiver engine with provided data, gateway port:%d, threads:%d", gatewayPort, numThreads);
        return DIVERT_ERROR_FAILED;
    }
    if( listenPort < 1 || listenPort > 65536 || gatewayHost == NULL || listenHost == NULL ) {
        logNativeMessageToGo(0, "Cannot initialize windiver engine with provided data, listen port:%d, gatewayHost:%s, listenHost:%s", 
            listenPort, gatewayHost ? gatewayHost : NULL, listenHost ? listenHost : NULL );
        return DIVERT_ERROR_FAILED;
    }

    logNativeMessageToGo(0, "Initializing windivert engine..." ); 
    InitializeSRWLock(&sharedRWLock);

    // The filter for windivert, captures outbound tcp packets which are not directed at the client listening port
    char filterOut[256] = "";
    snprintf(filterOut, 256, FILTER_OUTBOUND, listenPort);
    logNativeMessageToGo(0, "Filtering outbound with %s", filterOut);

    // Open Windivert engine
    diverterHandle = WinDivertOpen( filterOut, WINDIVERT_LAYER_NETWORK, 0, 0 );
    if (diverterHandle == INVALID_HANDLE_VALUE) {
        logNativeMessageToGo(0, "Could not initialize windivert engine, errorcode %d", GetLastError());
        return DIVERT_ERROR_NOTINITILIZED;
    }

    for( int i=0; i<MAX_THREADS; i++ ) {
        threadHandles[i] = INVALID_HANDLE_VALUE;
    }
    // Set the connections to initial state
    for( int i=0; i<65536; i++ ) {
        connectionsList[i].origSrcPort = 0;
        connectionsList[i].origDstPort = 0;
        connectionsList[i].state = STATE_CLOSED;
    }

    // Initialize the worker threads
    for( int i=0; i<numThreads; i++ ) {
        threadParameters* th = (threadParameters*)malloc( sizeof(threadParameters) );
        th->gatewayAddress = gatewayHost;
        th->gatewayPort = gatewayPort;
        th->listenAddress = listenHost;
        th->listenPort = listenPort;
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

/**
 * @brief Stops the worker threads and closes the divert engine
 * 
 * @return DIVERT_OK    if everything ok, an error otherwise
 */
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

/**
 * @brief Main dispatching routine for captured packets
 * 
 * @param lpParameter 
 * @return DWORD 
 */
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
        // Receive packets
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
            // Not handled
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

        dumpPacket(th->threadID, (PVOID)packet, packetLen );

        // creates the windivert device destination address
        memcpy(&send_addr, &recv_addr, sizeof(WINDIVERT_ADDRESS));

        if( localSrcPort != th->listenPort ) {
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
        send_addr.Impostor = 1; // this tells windivert that the packets was already redirected and avoids loops

        // Reparsing for updated data
        WinDivertHelperParsePacket(packet, packetLen, 
            &ip_header, &ipv6_header,
            NULL, NULL, NULL, 
            &tcp_header, 
            NULL, 
            &packetData, &packetDataLen,
            &next, &nextLen);

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

        // announce new packet
        logNativeMessageToGo(th->threadID,  "Redirected packet: %s:%d %s %s:%d (%d) [S:%d A:%d F:%d P:%d R:%d]", 
            src_str, ntohs(tcp_header->SrcPort), 
            (recv_addr.Outbound? "---->": "<----"),
            dst_str, ntohs(tcp_header->DstPort),
            ntohs(tcp_header->SeqNum),
            tcp_header->Syn, tcp_header->Ack,
            tcp_header->Fin, tcp_header->Psh,
            tcp_header->Rst );

        dumpPacket( th->threadID, (PVOID)packet, packetLen );

        // Calculates the new checksums
        if( WinDivertHelperCalcChecksums(packet, packetLen, &send_addr, 0) != TRUE ) 
        {
            logNativeMessageToGo(th->threadID, "Could not calculate checksum, dropping...");
            continue;
        }

        // Injects the modified packet to the network
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

    return DIVERT_OK;
}

/**
 * @brief Updates only the connection status atomically
 * 
 * @param port   Port of the connection
 * @param state  New status of the connection
 */
void atomicUpdateConnectionState(UINT port, int state) {
    if( port < 1 || port > 65536 ) {
        logNativeMessageToGo(0, "Invalid port provided, port:%d", port);
        return;
    }
    if( state < STATE_CLOSED || state >= STATE_MAX ) {
        logNativeMessageToGo(0, "Invalid connection state provided, state:%d", state);
        return;
    }

    if( !acquireLock(TRUE) ) {
        return;
    }
    connectionsList[ port ].state = state;
    releaseLock(TRUE);
}

/**
 * @brief Handles the redirect logic when packet is from the local machine to the remote server
 * 
 * eg. 54731 -> 8080 redirected as 54731 -> 9443 (connection index 54731)
 * 
 * Actually the redirection is pointed at the local client, substituting the destination port 
 * with the listening port of the client and also the address.
 * 
 * Once the connection is established in the client.go listener, then the QUIC header is changed
 * back to the remote destination so as to mantain the actual destination available.
 * 
 * Note on the recv_addr and send_addr: those are structures used by WinDivert to manage the
 * correct routing to the device driver indenpendently to the actual content of the packet.
 * 
 * @param th           Thread parameters structure pointer
 * @param ip_header    Parsed IPv4 header pointer to the packet
 * @param ipv6_header  Parsed IPv46 header pointer to the packet
 * @param tcp_header   Parsed TCP protocol header pointer to the packet
 * @param recv_addr    Received address from the windivert device
 * @param send_addr    Destination address from the windivert device
 * @return             TRUE if packet was actually redirected, FALSE if dropped
 */
BOOL handleLocalToServerPacket(
    threadParameters* th,
    PWINDIVERT_IPHDR ip_header,
    PWINDIVERT_IPV6HDR ipv6_header,
    PWINDIVERT_TCPHDR tcp_header,
    WINDIVERT_ADDRESS* recv_addr,
    WINDIVERT_ADDRESS* send_addr ) 
{
    // Critical section to acquire the current connection data
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
    // Handles the sequence by which the TCP handshake is expected
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
            if( tcp_header->Fin || tcp_header->Rst ) {
                logNativeMessageToGo(th->threadID,  "Reset/Fin received for port %d, closing connection", portSrcIdx);
                atomicUpdateConnectionState( portSrcIdx, STATE_WAIT );
            } else {
                logNativeMessageToGo(th->threadID,  "Out-of-sequence packet for handshake, dropping...");
                return FALSE;
            }

        case STATE_SYN_ACK:
            if( tcp_header->Fin || tcp_header->Rst ) {
                logNativeMessageToGo(th->threadID,  "Reset received for port %d, closing connection", portSrcIdx);
                atomicUpdateConnectionState( portSrcIdx, STATE_WAIT );
            } else if( !tcp_header->Ack ) {
                logNativeMessageToGo(th->threadID,  "Out-of-sequence packet for handshake, dropping...");
                return FALSE;
            }
            connState = STATE_OPEN;
            break;

        case STATE_WAIT:
            if( !(tcp_header->Ack) ) {
                logNativeMessageToGo(th->threadID,  "Out-of-sequence packet for handshake, dropping...");
                return FALSE;
            }
            logNativeMessageToGo(th->threadID,  "Reset received for port %d, closing connection", portSrcIdx);
            atomicUpdateConnectionState( portSrcIdx, STATE_CLOSED );
            break;

        case STATE_OPEN:
            if( tcp_header->Fin || tcp_header->Rst ) {
                logNativeMessageToGo(th->threadID,  "Reset received for port %d, closing connection", portSrcIdx);
                atomicUpdateConnectionState( portSrcIdx, STATE_WAIT );
            }
            break;
    }

    BOOL isIPV4 = ( ip_header != NULL ) ? TRUE : FALSE;
    BOOL isIPV6 = ( ipv6_header != NULL ) ? TRUE : FALSE;

    // update (and acquire lock) only if connection state has changed
    if( connState != connStatePrev ) {
        logNativeMessageToGo(th->threadID,  "NEW Connection state for port %d: %s", portSrcIdx, connStateToString(connState));

        // update state and set ports if necessary
        if( !acquireLock(TRUE) ) {
            logNativeMessageToGo(th->threadID,  "Lock acquire timeout, dropping packet...");
            return FALSE;
        }
        connectionsList[ portSrcIdx ].state = connState;
        if( connectionsList[ portSrcIdx ].origSrcPort == 0 ) {
            connectionsList[ portSrcIdx ].connectionIPV4 = isIPV4;
            connectionsList[ portSrcIdx ].connectionIPV6 = isIPV6;
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
    tcp_header->DstPort = htons(th->listenPort);

    if( isIPV4 )
    {
        UINT32 remote_addr;
        WinDivertHelperParseIPv4Address(th->listenAddress, &remote_addr);
        ip_header->DstAddr = htonl(remote_addr);
    }
    if( isIPV6 )
    {
        UINT32 remote_addr[4];
        WinDivertHelperParseIPv6Address(th->listenAddress, remote_addr);
        ipv6_header->DstAddr[0] = htonl(remote_addr[0]);
        ipv6_header->DstAddr[1] = htonl(remote_addr[1]);
        ipv6_header->DstAddr[2] = htonl(remote_addr[2]);
        ipv6_header->DstAddr[3] = htonl(remote_addr[3]);
    }

    // redirect data for windivert engine
    WinDivertHelperParseIPv4Address(th->listenAddress, send_addr->Flow.RemoteAddr);
    send_addr->Flow.RemotePort = th->listenPort;

    WinDivertHelperParseIPv4Address(th->listenAddress, send_addr->Socket.RemoteAddr);
    send_addr->Socket.RemotePort = th->listenPort;

    return TRUE;
}

/**
 * @brief Handles the redirect logic when packet is from the remote server to the local machine
 * 
 * eg. 9443 -> 54731 redirected as 8080 -> 54731 (connection index 54731)
 * 
 * In this case the change lies in the source port, which is changed with the original
 * destination port of the connection, this way the source software does not have to know 
 * about the presence of the diverter and the connection is transparent as if it was 
 * connecting directly to the original destination.
 * 
 * Note however that the connection is never established from the outside so only "return" packets
 * are allowed, no new connection can be opened.
 * 
 * Note on the recv_addr and send_addr: those are structures used by WinDivert to manage the
 * correct routing to the device driver indenpendently to the actual content of the packet.
 * 
 * @param th           Thread parameters structure pointer
 * @param ip_header    Parsed IPv4 header pointer to the packet
 * @param ipv6_header  Parsed IPv46 header pointer to the packet
 * @param tcp_header   Parsed TCP protocol header pointer to the packet
 * @param recv_addr    Received address from the windivert device
 * @param send_addr    Destination address from the windivert device
 * @return             TRUE if packet was actually redirected, FALSE if dropped
 */
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

    // Lock the connection in read and acquire the current state of the connection
    if( !acquireLock(FALSE) ) {
        logNativeMessageToGo(th->threadID,  "Lock acquire timeout, dropping packet...");
        return FALSE;
    }
    int portSrcIdx = (int)ntohs(tcp_header->DstPort);
    UINT connState = connectionsList[ portSrcIdx ].state;
    UINT connStatePrev = connState;

    UINT origSrcPort = connectionsList[ portSrcIdx ].origSrcPort;
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
    // Handle the packet in respect to the expected tcp handshake 
    logNativeMessageToGo(th->threadID,  "Connection state for port %d: %s", portSrcIdx, connStateToString(connState));
    switch( connState ) {
        case STATE_CLOSED:
            logNativeMessageToGo(th->threadID,  "Connection can only be established from local side, dropped");
            return FALSE;

        case STATE_SYN:
            if( tcp_header->Fin || tcp_header->Rst ) {
                logNativeMessageToGo(th->threadID,  "Reset received for port %d, closing connection", portSrcIdx);
                atomicUpdateConnectionState( portSrcIdx, STATE_WAIT );
            } else if( !(tcp_header->Syn && tcp_header->Ack) ) {
                logNativeMessageToGo(th->threadID,  "Out-of-sequence packet for handshake, dropping...");
                return FALSE;
            }
            connState = STATE_SYN_ACK;
            break;

        case STATE_SYN_ACK:
            if( tcp_header->Fin || tcp_header->Rst ) {
                logNativeMessageToGo(th->threadID,  "Reset received for port %d, closing connection", portSrcIdx);
                atomicUpdateConnectionState( portSrcIdx, STATE_WAIT );
            } else if( !(tcp_header->Ack) ) {
                logNativeMessageToGo(th->threadID,  "Out-of-sequence packet for handshake, dropping...");
                return FALSE;
            }
            break;

        case STATE_WAIT:
            if( !(tcp_header->Ack) ) {
                logNativeMessageToGo(th->threadID,  "Out-of-sequence packet for handshake, dropping...");
                return FALSE;
            }
            logNativeMessageToGo(th->threadID,  "Reset received for port %d, closing connection", portSrcIdx);
            atomicUpdateConnectionState( portSrcIdx, STATE_CLOSED );
            break;

        case STATE_OPEN:
            if( tcp_header->Fin || tcp_header->Rst ) {
                logNativeMessageToGo(th->threadID,  "Reset received for port %d, closing connection", portSrcIdx);
                atomicUpdateConnectionState( portSrcIdx, STATE_WAIT );
            }
    }

    // Only acquire the lock if state was changed
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
    tcp_header->DstPort = htons(origSrcPort);

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

    send_addr->Flow.RemotePort = htons(origSrcPort);
    send_addr->Socket.RemotePort = htons(origSrcPort);
    send_addr->Flow.LocalPort = htons(origDstPort);
    send_addr->Socket.LocalPort = htons(origDstPort);

    return TRUE;
}

// assumes network (big-endian) byte order
void ipV4PackedToUnpackedNetworkByteOrder( UINT32 packed, UINT32* unpacked ) {
    UINT32 local = ntohl(packed);

    unpacked[0] = (UINT8)((local & (0xFF000000)) >> 24);
    unpacked[1] = (UINT8)((local & (0x00FF0000)) >> 16);
    unpacked[2] = (UINT8)((local & (0x0000FF00)) >> 8);
    unpacked[3] = (UINT8)(local & (0x000000FF));
}

/**
 * @brief Dumps the contents of the packet to the go log
 * 
 * @param thid        Thread ID
 * @param packetData  Data of the packet
 * @param len         Length of the data
 */
void dumpPacket( int thid, PVOID packetData, UINT len ) {
    if( !diveterMessagesEnabledToGo || packetData == NULL || len < 1 )
        return;

    char linebuff[2048] = "";
    char *currbuff = linebuff;

    currbuff += sprintf(currbuff, "%06X ", 0);
    for( int i=0; i<len; i++ ) {
        if( i > 0 && i % 16 == 0 ) {
            currbuff++;
            (*currbuff) = '\0';

            logNativeMessageToGo( thid, "%s", linebuff );

            currbuff = linebuff;
            (*currbuff) = '\0';

            currbuff += sprintf(currbuff, "%06X ", i-16);
        }
            
        currbuff += sprintf(currbuff, "%02X ", ((unsigned char*)packetData)[i]);
    }
    if( len == 0 || len % 16 != 0 ) {
        currbuff++;
        (*currbuff) = '\0';

        logNativeMessageToGo( thid, "%s", linebuff );
    }
}

/**
 * @brief Prints messages to the go log printf-style
 * 
 * @param thid    Thread ID
 * @param format  String printf-style to print
 * @param ...     Successive parameters are treated as in printf
 */
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

/**
 * @brief Enables or disabled the loggin of diverter to go
 * 
 * Please be sure to use this only in a debug context, it has very
 * heavy performance penalty
 * 
 * @param enabled   FALSE messages are ignored, TRUE messages are printed
 */
void EnableMessageOutputToGo( int enabled ) {
    if( enabled > 0 )
        enabled = TRUE;
    if( enabled < 0 )
        enabled = FALSE;
    diveterMessagesEnabledToGo = enabled;
}

/**
 * @brief Acquires a shared lock with a timeout
 * 
 * Be sure to always have matching TRUE lock -> TRUE unlock and
 * FALSE lock -> FALSE unlock or deadlock may occur!
 * 
 * @param exclusive   TRUE = write lock, FALSE = read lock
 * @return BOOLEAN    TRUE = lock acquired, FALSE = lock timeout
 */
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

/**
 * @brief See acquireLock
 * 
 * @param exclusive TRUE = write lock, FALSE = read lock
 */
void releaseLock( BOOLEAN exclusive ) {
    if( exclusive ) {
        ReleaseSRWLockExclusive(&sharedRWLock);
        return;
    }

    ReleaseSRWLockShared(&sharedRWLock);
}

/**
 * @brief Returns the string representation of the state value
 * 
 * @param state         State provided
 * @return const char*  String value of the input state, "unknown" if invalid value
 */
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
        case STATE_WAIT:
            return "waiting ack";
    }
    return "unknown";
}

/**
 * @brief Recovers the data about the connection in a synchronized way
 * 
 * @param sourcePort      Port of the connection
 * @param origSrcPort     Original source port of the connection
 * @param origDstPort     Original destination port of the connection
 * @param origSrcAddress  Original source address of the connection
 * @param origDstAddress  Original destination address of the connection
 * @return int            DIVERT_OK if ok, error otherwise
 */
int  GetConnectionData( UINT sourcePort, UINT* origSrcPort, UINT* origDstPort, 
                               char* origSrcAddress, char* origDstAddress )
{
    if( sourcePort < 1 || sourcePort > 65536 ) {
        logNativeMessageToGo(0, "Invalid port:%d", sourcePort);
        return DIVERT_ERROR_FAILED;
    }

    if( !acquireLock( FALSE ) )
        return DIVERT_ERROR_FAILED;

    connection* c = &connectionsList[sourcePort];

    if( c->state == STATE_CLOSED ) {
        logNativeMessageToGo(0, "Connection CLOSED on port:%d", sourcePort);
        return DIVERT_ERROR_NOT_OPEN;
    }

    if( origSrcPort != NULL )
        *origSrcPort = c->origSrcPort;
    if( origDstPort != NULL )
        *origDstPort = c->origDstPort;
    if( origSrcAddress != NULL ) {
        if( c->connectionIPV4 )
            WinDivertHelperFormatIPv4Address(ntohl(c->origSrcAddress), origSrcAddress, 64);
        if( c->connectionIPV6 ) {
            UINT32 tmp[4];
            tmp[0] = ntohs(c->origSrcAddressV6[0]);
            tmp[1] = ntohs(c->origSrcAddressV6[1]);
            tmp[2] = ntohs(c->origSrcAddressV6[2]);
            tmp[3] = ntohs(c->origSrcAddressV6[3]);
            WinDivertHelperFormatIPv6Address(tmp, origSrcAddress, 64);
        }
    }
    if( origDstAddress != NULL ) {
        if( c->connectionIPV4 )
            WinDivertHelperFormatIPv4Address(ntohl(c->origDstAddress), origDstAddress, 64);
        if( c->connectionIPV6 ) {
            UINT32 tmp[4];
            tmp[0] = ntohs(c->origDstAddressV6[0]);
            tmp[1] = ntohs(c->origDstAddressV6[1]);
            tmp[2] = ntohs(c->origDstAddressV6[2]);
            tmp[3] = ntohs(c->origDstAddressV6[3]);
            WinDivertHelperFormatIPv6Address(tmp, origDstAddress, 64);
        }
    }

    releaseLock( FALSE );
    return DIVERT_OK;
}

#endif
