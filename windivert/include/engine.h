#pragma once

//#define FILTER_OUTBOUND "!impostor and tcp and tcp.DstPort!=%d"
#define FILTER_OUTBOUND "!impostor and tcp and tcp.DstPort!=%d && tcp.SrcPort != %d && tcp.DstPort != %d"

#define MAXBUF            WINDIVERT_MTU_MAX
#define INET6_ADDRSTRLEN  45
#define MAX_THREADS       8

#define ntohs(x) WinDivertHelperNtohs(x)
#define ntohl(x) WinDivertHelperNtohl(x)
#define htons(x) WinDivertHelperHtons(x)
#define htonl(x) WinDivertHelperHtonl(x)

enum {
  STATE_CLOSED = 0,
  STATE_SYN = 1,
  STATE_SYN_ACK = 2,
  STATE_OPEN = 3,
  STATE_WAIT = 4,
};

typedef struct
{
    WINDIVERT_IPHDR ip;
    WINDIVERT_TCPHDR tcp;
    unsigned char data[MAXBUF];
} TCPPACKET;

typedef struct
{
    WINDIVERT_IPV6HDR ipv6;
    WINDIVERT_TCPHDR tcp;
    unsigned char data[MAXBUF];
} TCPV6PACKET;

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
    int threadID;
    int gatewayPort;
    char* gatewayAddress;
} threadParameters;

DWORD WINAPI dispatchDivertedOutboundPackets(LPVOID lpParameter);

BOOLEAN     acquireLock(BOOLEAN exclusive);
void        releaseLock(BOOLEAN exclusive);

void         logNativeMessageToGo(int thid, const char *format...);
const char*  connStateToString(UINT state);
void         dumpPacket( PVOID packetData, UINT len );

void ipV4PackedToUnpackedNetworkByteOrder( UINT32 packed, UINT32* unpacked );

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
