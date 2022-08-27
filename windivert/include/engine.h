#pragma once

#define FILTER_OUTBOUND "!impostor and ip.SrcAddr!=127.0.0.1 and ip.DstAddr!=127.0.0.1 and tcp and tcp.DstPort!=%d"

#define MAXBUF            WINDIVERT_MTU_MAX
#define INET6_ADDRSTRLEN  45
#define MAX_THREADS       8

#define ntohs(x) WinDivertHelperNtohs(x)
#define ntohl(x) WinDivertHelperNtohl(x)
#define htons(x) WinDivertHelperHtons(x)
#define htonl(x) WinDivertHelperHtonl(x)

enum {
  STATE_CLOSED = 0,   //!< Connection is closed
  STATE_SYN = 1,      //!< Connection received SYN, awaits for SYN ACK
  STATE_SYN_ACK = 2,  //!< Connection received SYN ACK, awaits for ACK
  STATE_OPEN = 3,     //!< Connection is open for push / ack packets
  STATE_WAIT = 4,     //!< FIN or RST packet was received, waiting for ACK
  STATE_MAX           //!< Max status for checks
};

typedef struct
{
    UINT   state;
    UINT   origSrcPort;
    UINT   origDstPort;
    BOOL   connectionIPV4;
    BOOL   connectionIPV6;
    UINT32  origSrcAddress;
    UINT32  origDstAddress;
    UINT32  origSrcAddressV6[4];
    UINT32  origDstAddressV6[4];
} connection;

typedef struct
{
    int   threadID;
    int   gatewayPort;
    char* gatewayAddress;
    int   listenPort;
    char* listenAddress;
} threadParameters;

DWORD WINAPI dispatchDivertedOutboundPackets(LPVOID lpParameter);

BOOLEAN     acquireLock(BOOLEAN exclusive);
void        releaseLock(BOOLEAN exclusive);

void         logNativeMessageToGo(int thid, const char *format...);
const char*  connStateToString(UINT state);
void         dumpPacket( int thid, PVOID packetData, UINT len );
void         atomicUpdateConnectionState(UINT port, int state);
void         ipV4PackedToUnpackedNetworkByteOrder( UINT32 packed, UINT32* unpacked );

BOOL handleLocalToServerPacket(
    threadParameters* th,
    PWINDIVERT_IPHDR ip_header,
    PWINDIVERT_IPV6HDR ipv6_header,
    PWINDIVERT_TCPHDR tcp_header,
    WINDIVERT_ADDRESS* recv_addr,
    WINDIVERT_ADDRESS* send_addr );

BOOL handleServerToLocalPacket(
    threadParameters* th,
    PWINDIVERT_IPHDR ip_header,
    PWINDIVERT_IPV6HDR ipv6_header,
    PWINDIVERT_TCPHDR tcp_header,
    WINDIVERT_ADDRESS* recv_addr,
    WINDIVERT_ADDRESS* send_addr );
