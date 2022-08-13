package shared

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jackpal/gateway"
)

type QuicConfig struct {
	AckElicitingPacketsBeforeAck   int
	AckDecimationDenominator       int
	InitialCongestionWindowPackets int
	MultiStream                    bool
	VarAckDelay                    float64
	MaxAckDelay                    int //in miliseconds, used to determine if decimating
	MinReceivedBeforeAckDecimation int
	ClientFlag                     bool
	GatewayIP                      string
	GatewayPort                    int
	GatewayAPIPort                 int
	ListenIP                       string
	ListenPort                     int
	WinDivertThreads               int
	Verbose                        bool
}

const (
	ACK_FLAG = "acks"
)

var (
	QuicConfiguration       QuicConfig
	defaultListeningAddress string
)

func ParseFlags(args []string) {
	log.Println("ARGS:", strings.Join(args, " "))

	ackElicitingFlag := flag.Int("acks", 10, "Number of acks to bundle")
	ackDecimationFlag := flag.Int("decimate", 4, "Denominator of Ack Decimation Ratio")
	congestionWindowFlag := flag.Int("congestion", 4, "Number of QUIC packets for initial congestion window")
	multiStreamFlag := flag.Bool("multistream", true, "Enable multiplexed QUIC streams inside a single session")
	maxAckDelayFlag := flag.Int("ackDelay", 25, "Maximum number of miliseconds to hold back an ack for decimation")
	varAckDelayFlag := flag.Float64("varAckDelay", 0.25, "Variable number of miliseconds to hold back an ack for decimation, as multiple of RTT")
	minReceivedBeforeAckDecimationFlag := flag.Int("minBeforeDecimation", 100, "Minimum number of packets before initiating ack decimation")
	clientFlag := flag.Bool("client", false, "a bool")
	gatewayHostFlag := flag.String("gateway", "198.18.0.254", "IP address of gateway running qpep server")
	gatewayPortFlag := flag.Int("port", 443, "Port of gateway running qpep server")
	gatewayAPIPortFlag := flag.Int("apiport", 444, "IP address of gateway running qpep server")
	listenHostFlag := flag.String("listenaddress", "0.0.0.0", "IP listen address of qpep client")
	listenPortFlag := flag.Int("listenport", 9443, "Listen Port of qpep client")
	winDiverterThreads := flag.Int("threads", 1, "Worker threads for windivert engine (min 1, max 8)")
	verbose := flag.Bool("verbose", false, "Outputs data about diverted connections for debug")

	flag.CommandLine.Parse(args[1:])
	if !flag.Parsed() {
		flag.Usage()
		os.Exit(1)
	}

	QuicConfiguration = QuicConfig{
		AckElicitingPacketsBeforeAck:   *ackElicitingFlag,
		AckDecimationDenominator:       *ackDecimationFlag,
		InitialCongestionWindowPackets: *congestionWindowFlag,
		MultiStream:                    *multiStreamFlag,
		MaxAckDelay:                    *maxAckDelayFlag,
		VarAckDelay:                    *varAckDelayFlag,
		MinReceivedBeforeAckDecimation: *minReceivedBeforeAckDecimationFlag,
		ClientFlag:                     *clientFlag,
		GatewayIP:                      *gatewayHostFlag,
		GatewayPort:                    *gatewayPortFlag,
		GatewayAPIPort:                 *gatewayAPIPortFlag,
		ListenIP:                       *listenHostFlag,
		ListenPort:                     *listenPortFlag,
		WinDivertThreads:               *winDiverterThreads,
		Verbose:                        *verbose,
	}

	data, _ := json.MarshalIndent(QuicConfiguration, "", " ")
	log.Printf("%v\n", string(data))
}

func GetDefaultLanListeningAddress(currentAddress string) string {
	if len(defaultListeningAddress) > 0 {
		return defaultListeningAddress
	}

	if !strings.HasPrefix(currentAddress, "0.") && !strings.HasPrefix(currentAddress, "127.") {
		return currentAddress
	}

	log.Printf("WARNING: Detected invalid listening ip address, trying to autodetect the default route...\n")

	defaultIP, err := gateway.DiscoverInterface()
	if err != nil {
		panic(fmt.Sprint("PANIC: Could not discover default lan address and the requested one is not suitable, error: %v\n", err))
		return currentAddress
	}

	defaultListeningAddress = defaultIP.String()
	log.Printf("Found default ip address: %s\n", defaultListeningAddress)
	return defaultListeningAddress
}
