package client

import (
	"crypto/tls"
	"io"
	"log"
	"net"
	"runtime/debug"
	"strconv"
	"sync"
	"time"

	"github.com/lucas-clemente/quic-go"
	"github.com/parvit/qpep/api"
	"github.com/parvit/qpep/shared"
	"github.com/parvit/qpep/windivert"
	"golang.org/x/net/context"
)

var (
	proxyListener       net.Listener
	ClientConfiguration = ClientConfig{
		ListenHost: "0.0.0.0", ListenPort: 9443,
		GatewayHost: "198.56.1.10", GatewayPort: 443,
		QuicStreamTimeout: 2, MultiStream: shared.QuicConfiguration.MultiStream,
		ConnectionRetries: 3,
		IdleTimeout:       time.Duration(300) * time.Second,
		WinDivertThreads:  1,
		Verbose:           false,
	}
	quicSession             quic.Session
	QuicClientConfiguration = quic.Config{
		MaxIncomingStreams: 40000,
	}
)

type ClientConfig struct {
	ListenHost        string
	ListenPort        int
	GatewayHost       string
	GatewayPort       int
	APIPort           int
	QuicStreamTimeout int
	MultiStream       bool
	IdleTimeout       time.Duration
	ConnectionRetries int
	WinDivertThreads  int
	Verbose           bool
}

func RunClient(ctx context.Context) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("PANIC: %v", err)
			debug.PrintStack()
		}
		if proxyListener != nil {
			proxyListener.Close()
		}
	}()
	log.Println("Starting TCP-QPEP Tunnel Listener")

	// update configuration from flags
	ClientConfiguration.GatewayHost = shared.QuicConfiguration.GatewayIP
	ClientConfiguration.GatewayPort = shared.QuicConfiguration.GatewayPort
	ClientConfiguration.APIPort = shared.QuicConfiguration.GatewayAPIPort
	ClientConfiguration.ListenHost = shared.QuicConfiguration.ListenIP
	ClientConfiguration.ListenPort = shared.QuicConfiguration.ListenPort
	ClientConfiguration.MultiStream = shared.QuicConfiguration.MultiStream
	ClientConfiguration.WinDivertThreads = shared.QuicConfiguration.WinDivertThreads
	ClientConfiguration.Verbose = shared.QuicConfiguration.Verbose

	log.Printf("Binding to TCP %s:%d", ClientConfiguration.ListenHost, ClientConfiguration.ListenPort)
	var err error
	proxyListener, err = NewClientProxyListener("tcp", &net.TCPAddr{
		IP:   net.ParseIP(ClientConfiguration.ListenHost),
		Port: ClientConfiguration.ListenPort,
	})
	if err != nil {
		log.Printf("Encountered error when binding client proxy listener: %s", err)
		return
	}

	go ListenTCPConn()

	for {
		select {
		case <-ctx.Done():
			proxyListener.Close()
			return
		case <-time.After(1 * time.Second):
			apiStatusCheck()
			continue
		}
	}
}

func ListenTCPConn() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("PANIC: %v", err)
			debug.PrintStack()
		}
	}()
	for {
		conn, err := proxyListener.Accept()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Temporary() {
				log.Printf("Temporary error when accepting connection: %s", netErr)
			}
			log.Printf("Unrecoverable error while accepting connection: %s", err)
			return
		}

		go handleTCPConn(conn)
	}
}

func handleTCPConn(tcpConn net.Conn) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("PANIC: %v", err)
			debug.PrintStack()
		}
	}()
	log.Printf("Accepting TCP connection from %s with destination of %s", tcpConn.RemoteAddr().String(), tcpConn.LocalAddr().String())
	defer tcpConn.Close()
	var quicStream quic.Stream = nil
	// if we allow for multiple streams in a session, lets try and open on the existing session
	if ClientConfiguration.MultiStream {
		//if we have already opened a quic session, lets check if we've expired our stream
		if quicSession != nil {
			var err error
			log.Printf("Trying to open on existing session")
			quicStream, err = quicSession.OpenStream()
			// if we weren't able to open a quicStream on that session (usually inactivity timeout), we can try to open a new session
			if err != nil {
				log.Printf("Unable to open new stream on existing QUIC session: %s\n", err)
				quicStream = nil
			} else {
				log.Printf("Opened a new stream: %d", quicStream.StreamID())
			}
		}
	}
	// if we haven't opened a stream from multistream, we can open one with a new session
	if quicStream == nil {
		// open a new quicSession (with all the TLS jazz)
		var err error
		quicSession, err = openQuicSession()
		// if we were unable to open a quic session, drop the TCP connection with RST
		if err != nil {
			return
		}

		//Open a stream to send data on this new session
		quicStream, err = quicSession.OpenStreamSync(context.Background())
		// if we cannot open a stream on this session, send a TCP RST and let the client decide to try again
		if err != nil {
			log.Printf("Unable to open QUIC stream: %s\n", err)
			return
		}
	}
	defer quicStream.Close()

	//We want to wait for both the upstream and downstream to finish so we'll set a wait group for the threads
	var streamWait sync.WaitGroup
	streamWait.Add(2)

	//Set our custom header to the QUIC session so the server can generate the correct TCP handshake on the other side
	sessionHeader := shared.QpepHeader{
		SourceAddr: tcpConn.RemoteAddr().(*net.TCPAddr),
		DestAddr:   tcpConn.LocalAddr().(*net.TCPAddr),
	}

	diverted, srcPort, dstPort, srcAddress, dstAddress := windivert.GetConnectionStateData(sessionHeader.SourceAddr.Port)
	if diverted == windivert.DIVERT_OK {
		log.Printf("Diverted connection: %v:%v %v:%v", srcAddress, srcPort, dstAddress, dstPort)

		sessionHeader.SourceAddr = &net.TCPAddr{
			IP:   net.ParseIP(srcAddress),
			Port: srcPort,
		}
		sessionHeader.DestAddr = &net.TCPAddr{
			IP:   net.ParseIP(dstAddress),
			Port: dstPort,
		}
	}

	log.Printf("Sending QUIC header to server, SourceAddr: %v / DestAddr: %v", sessionHeader.SourceAddr, sessionHeader.DestAddr)

	_, err := quicStream.Write(sessionHeader.ToBytes())
	if err != nil {
		log.Printf("Error writing to quic stream: %s", err.Error())
	}

	streamQUICtoTCP := func(dst *net.TCPConn, src quic.Stream) {
		_, err := io.Copy(dst, src)
		dst.SetLinger(3)
		dst.Close()
		//src.CancelRead(1)
		//src.Close()
		if err != nil {
			log.Printf("Error on Copy %s", err)
		}
		streamWait.Done()
	}

	streamTCPtoQUIC := func(dst quic.Stream, src *net.TCPConn) {
		_, err := io.Copy(dst, src)
		src.SetLinger(3)
		src.Close()
		//src.CloseWrite()
		//dst.CancelWrite(1)
		//dst.Close()
		if err != nil {
			log.Printf("Error on Copy %s", err)
		}
		streamWait.Done()
	}

	//Proxy all stream content from quic to TCP and from TCP to quic
	go streamTCPtoQUIC(quicStream, tcpConn.(*net.TCPConn))
	go streamQUICtoTCP(tcpConn.(*net.TCPConn), quicStream)

	//we exit (and close the TCP connection) once both streams are done copying
	streamWait.Wait()
	quicStream.Close()
	log.Printf("Done sending data on %d", quicStream.StreamID())
}

func openQuicSession() (quic.Session, error) {
	var err error
	var session quic.Session
	tlsConf := &tls.Config{InsecureSkipVerify: true, NextProtos: []string{"qpep"}}
	gatewayPath := ClientConfiguration.GatewayHost + ":" + strconv.Itoa(ClientConfiguration.GatewayPort)
	quicClientConfig := QuicClientConfiguration
	log.Printf("Dialing QUIC Session: %s\n", gatewayPath)
	for i := 0; i < ClientConfiguration.ConnectionRetries; i++ {
		session, err = quic.DialAddr(gatewayPath, tlsConf, &quicClientConfig)
		if err == nil {
			return session, nil
		} else {
			log.Printf("Failed to Open QUIC Session: %s\n    Retrying...\n", err)
		}
	}

	log.Printf("Max Retries Exceeded. Unable to Open QUIC Session: %s\n", err)
	return nil, err
}

func apiStatusCheck() {
	localAddr := ClientConfiguration.ListenHost
	apiAddr := ClientConfiguration.GatewayHost
	apiPort := ClientConfiguration.APIPort
	if response := api.RequestEcho(localAddr, apiAddr, apiPort); response != nil {
		log.Printf("Gateway Echo OK\n")
		return
	}
	log.Printf("Gateway Echo FAILED\n")
}
