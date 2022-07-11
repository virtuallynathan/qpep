package server

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"math/big"
	"net"
	"runtime/debug"
	"strconv"
	"sync"
	"time"

	"github.com/parvit/qpep/client"
	"github.com/parvit/qpep/shared"

	"github.com/lucas-clemente/quic-go"
)

var (
	ServerConfiguration = ServerConfig{
		ListenHost: "127.0.0.1",
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
			log.Printf("PANIC: %v", err)
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
	log.Printf("Opening QPEP Server on: %s", listenAddr)
	var err error
	quicListener, err = quic.ListenAddr(listenAddr, generateTLSConfig(), &client.QuicClientConfiguration)
	if err != nil {
		log.Printf("Encountered error while binding QUIC listener: %s", err)
		return
	}
	defer quicListener.Close()

	go ListenQuicSession()

	for {
		select {
		case <-ctx.Done():
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
			log.Printf("PANIC: %v", err)
			debug.PrintStack()
		}
	}()
	for {
		var err error
		quicSession, err = quicListener.Accept(context.Background())
		if err != nil {
			log.Printf("Unrecoverable error while accepting QUIC session: %s", err)
			return
		}
		Statistics.Increment(TOTAL_CONNECTIONS)
		go ListenQuicConn(quicSession)
	}
}

func ListenQuicConn(quicSession quic.Session) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("PANIC: %v", err)
			debug.PrintStack()
		}
		Statistics.Decrement(TOTAL_CONNECTIONS)
	}()
	for {
		stream, err := quicSession.AcceptStream(context.Background())
		if err != nil {
			if err.Error() != "NO_ERROR: No recent network activity" {
				log.Printf("Unrecoverable error while accepting QUIC stream: %s", err)
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
			log.Printf("PANIC: %v", err)
			debug.PrintStack()
		}
	}()
	qpepHeader, err := shared.GetQpepHeader(stream)
	if err != nil {
		log.Printf("Unable to find QPEP header: %s", err)
		return
	}
	go handleTCPConn(stream, qpepHeader)
}

func handleTCPConn(stream quic.Stream, qpepHeader shared.QpepHeader) {
	remoteConnIP := fmt.Sprintf(QUIC_CONN, qpepHeader.SourceAddr.IP)

	defer func() {
		if err := recover(); err != nil {
			log.Printf("PANIC: %v", err)
			debug.PrintStack()
		}
		Statistics.Decrement(remoteConnIP)
	}()

	Statistics.Increment(remoteConnIP)

	log.Printf("Opening TCP Connection to %s\n", qpepHeader.DestAddr.String())
	tcpConn, err := net.DialTimeout("tcp", qpepHeader.DestAddr.String(), time.Duration(10)*time.Second)
	if err != nil {
		log.Printf("Unable to open TCP connection from QPEP stream: %s", err)
		return
	}
	log.Printf("Opened TCP Conn %s -> %s\n", qpepHeader.SourceAddr.String(), qpepHeader.DestAddr.String())

	var streamWait sync.WaitGroup
	streamWait.Add(2)
	streamQUICtoTCP := func(dst *net.TCPConn, src quic.Stream) {
		_, err = io.Copy(dst, src)
		err1 := dst.SetLinger(3)
		if err1 != nil {
			log.Printf("error on setLinger: %s", err1)
		}
		dst.Close()
		if err != nil {
			log.Printf("Error on Copy %s", err)
		}
		streamWait.Done()
	}
	streamTCPtoQUIC := func(dst quic.Stream, src *net.TCPConn) {
		_, err = io.Copy(dst, src)
		log.Printf("Finished Copying TCP Conn %s->%s", src.LocalAddr().String(), src.RemoteAddr().String())
		err1 := src.SetLinger(3)
		if err1 != nil {
			log.Printf("error on setLinger: %s", err1)
		}
		src.Close()
		if err != nil {
			log.Printf("Error on Copy %s", err)
		}
		streamWait.Done()
	}

	go streamQUICtoTCP(tcpConn.(*net.TCPConn), stream)
	go streamTCPtoQUIC(stream, tcpConn.(*net.TCPConn))

	//we exit (and close the TCP connection) once both streams are done copying
	streamWait.Wait()
	stream.CancelRead(0)
	stream.CancelWrite(0)
	log.Printf("Closing TCP Conn %s->%s", tcpConn.LocalAddr().String(), tcpConn.RemoteAddr().String())
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
