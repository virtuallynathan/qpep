package server

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"io"
	"log"
	"math/big"
	"net"
	"runtime/debug"
	"strconv"
	"sync"
	"time"

	"github.com/parvit/qpep/api"
	"github.com/parvit/qpep/client"
	"github.com/parvit/qpep/shared"

	"github.com/lucas-clemente/quic-go"
)

const (
	INITIAL_BUFF_SIZE = int64(4096)
)

var (
	ServerConfiguration = ServerConfig{
		ListenHost: "0.0.0.0",
		ListenPort: 443,
		APIPort:    444,
	}
	quicListener quic.Listener
	quicSession  quic.Session
)

type ServerConfig struct {
	ListenHost string
	ListenPort int
	APIPort    int
}

func RunServer(ctx context.Context) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("PANIC: %v\n", err)
			debug.PrintStack()
		}
		if quicListener != nil {
			quicListener.Close()
		}
	}()

	// update configuration from flags
	ServerConfiguration.ListenHost = shared.QuicConfiguration.ListenIP
	ServerConfiguration.ListenPort = shared.QuicConfiguration.ListenPort
	ServerConfiguration.APIPort = shared.QuicConfiguration.GatewayAPIPort

	listenAddr := ServerConfiguration.ListenHost + ":" + strconv.Itoa(ServerConfiguration.ListenPort)
	log.Printf("Opening QPEP Server on: %s\n", listenAddr)
	var err error
	quicListener, err = quic.ListenAddr(listenAddr, generateTLSConfig(), &client.QuicClientConfiguration)
	if err != nil {
		log.Printf("Encountered error while binding QUIC listener: %s\n", err)
		return
	}
	defer quicListener.Close()

	go ListenQuicSession()

	ctxPerfWatcher, perfWatcherCancel := context.WithCancel(context.Background())
	go performanceWatcher(ctxPerfWatcher)

	for {
		select {
		case <-ctx.Done():
			perfWatcherCancel()
			quicListener.Close()
			return
		case <-time.After(10 * time.Millisecond):
			continue
		}
	}
}

func ListenQuicSession() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("PANIC: %v\n", err)
			debug.PrintStack()
		}
	}()
	for {
		var err error
		quicSession, err = quicListener.Accept(context.Background())
		if err != nil {
			log.Printf("Unrecoverable error while accepting QUIC session: %s\n", err)
			return
		}
		go ListenQuicConn(quicSession)
	}
}

func ListenQuicConn(quicSession quic.Session) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("PANIC: %v\n", err)
			debug.PrintStack()
		}
	}()
	for {
		stream, err := quicSession.AcceptStream(context.Background())
		if err != nil {
			if err.Error() != "NO_ERROR: No recent network activity" {
				log.Printf("Unrecoverable error while accepting QUIC stream: %s\n", err)
			}
			return
		}
		log.Printf("Opening QUIC StreamID: %d\n", stream.StreamID())

		go HandleQuicStream(stream)
	}
}

func HandleQuicStream(stream quic.Stream) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("PANIC: %v\n", err)
			debug.PrintStack()
		}
	}()
	qpepHeader, err := shared.GetQpepHeader(stream)
	if err != nil {
		log.Printf("Unable to find QPEP header: %s\n", err)
		return
	}
	go handleTCPConn(stream, qpepHeader)
}

func handleTCPConn(stream quic.Stream, qpepHeader shared.QpepHeader) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("PANIC: %v\n", err)
			debug.PrintStack()
		}
	}()

	timeOut := time.Duration(10) * time.Second

	log.Printf("Opening TCP Connection to %s, from %s\n", qpepHeader.DestAddr, qpepHeader.SourceAddr)
	tcpConn, err := net.DialTimeout("tcp", qpepHeader.DestAddr.String(), timeOut)
	if err != nil {
		log.Printf("Unable to open TCP connection from QPEP stream: %s\n", err)
		return
	}
	log.Printf("Opened TCP Conn %s -> %s\n", qpepHeader.SourceAddr, qpepHeader.DestAddr)

	trackedAddress := qpepHeader.SourceAddr.IP.String()
	proxyAddress := tcpConn.(*net.TCPConn).LocalAddr().String()

	api.Statistics.IncrementCounter(1.0, api.TOTAL_CONNECTIONS)
	api.Statistics.IncrementCounter(1.0, api.PERF_CONN, trackedAddress)
	defer func() {
		api.Statistics.DecrementCounter(1.0, api.PERF_CONN, trackedAddress)
		api.Statistics.DecrementCounter(1.0, api.TOTAL_CONNECTIONS)
	}()

	tcpConn.SetReadDeadline(time.Now().Add(timeOut))
	tcpConn.SetWriteDeadline(time.Now().Add(timeOut))
	stream.SetReadDeadline(time.Now().Add(timeOut))
	stream.SetWriteDeadline(time.Now().Add(timeOut))

	var streamWait sync.WaitGroup
	streamWait.Add(2)
	streamQUICtoTCP := func(dst *net.TCPConn, src quic.Stream) {
		defer func() {
			_ = recover()

			api.Statistics.DeleteMappedAddress(proxyAddress)
			dst.Close()
			streamWait.Done()
		}()

		api.Statistics.SetMappedAddress(proxyAddress, trackedAddress)

		err1 := dst.SetLinger(3)
		if err1 != nil {
			log.Printf("error on setLinger: %s\n", err1)
		}

		var buffSize = INITIAL_BUFF_SIZE
		for {
			written, err := io.Copy(dst, io.LimitReader(src, buffSize))
			if err != nil || written == 0 {
				log.Printf("Error on Copy %s\n", err)
				break
			}

			api.Statistics.IncrementCounter(float64(written), api.PERF_DW_COUNT, trackedAddress)
			buffSize = int64(written * 2)
			if buffSize < INITIAL_BUFF_SIZE {
				buffSize = INITIAL_BUFF_SIZE
			}
		}
		log.Printf("Finished Copying Stream ID %d, TCP Conn %s->%s\n", src.StreamID(), dst.LocalAddr().String(), dst.RemoteAddr().String())
	}
	streamTCPtoQUIC := func(dst quic.Stream, src *net.TCPConn) {
		defer func() {
			_ = recover()

			src.Close()
			streamWait.Done()
		}()

		err1 := src.SetLinger(3)
		if err1 != nil {
			log.Printf("error on setLinger: %s\n", err1)
		}

		var buffSize = INITIAL_BUFF_SIZE
		for {
			written, err := io.Copy(dst, io.LimitReader(src, INITIAL_BUFF_SIZE))
			if err != nil || written == 0 {
				log.Printf("Error on Copy %s\n", err)
				break
			}

			api.Statistics.IncrementCounter(float64(written), api.PERF_UP_COUNT, trackedAddress)
			buffSize = int64(written * 2)
			if buffSize < INITIAL_BUFF_SIZE {
				buffSize = INITIAL_BUFF_SIZE
			}
		}
		log.Printf("Finished Copying TCP Conn %s->%s, Stream ID %d\n", src.LocalAddr().String(), src.RemoteAddr().String(), dst.StreamID())
	}

	go streamQUICtoTCP(tcpConn.(*net.TCPConn), stream)
	go streamTCPtoQUIC(stream, tcpConn.(*net.TCPConn))

	//we exit (and close the TCP connection) once both streams are done copying
	streamWait.Wait()

	stream.CancelRead(0)
	stream.CancelWrite(0)
	log.Printf("Closing TCP Conn %s->%s\n", tcpConn.LocalAddr().String(), tcpConn.RemoteAddr().String())
}

func generateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"qpep"},
	}
}

func performanceWatcher(ctx context.Context) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("PANIC: %v\n", err)
			debug.PrintStack()
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(1 * time.Second):
			hosts := api.Statistics.GetHosts()

			for _, host := range hosts {
				// load the current count and reset it atomically (so there's no race condition)
				dwCount := api.Statistics.GetCounterAndClear(api.PERF_DW_COUNT, host)
				upCount := api.Statistics.GetCounterAndClear(api.PERF_UP_COUNT, host)
				if dwCount < 0.0 || upCount < 0.0 {
					continue
				}

				// update the speeds
				api.Statistics.SetCounter(dwCount/1024.0, api.PERF_DW_SPEED, host)
				api.Statistics.SetCounter(upCount/1024.0, api.PERF_UP_SPEED, host)

				// update the totals for the client
				api.Statistics.IncrementCounter(dwCount, api.PERF_DW_TOTAL, host)
				api.Statistics.IncrementCounter(upCount, api.PERF_UP_TOTAL, host)
			}
		}
	}
}
