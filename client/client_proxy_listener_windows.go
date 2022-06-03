//go:build windows
// +build windows

package client

import (
	"net"
)

type ClientProxyListener struct {
	base net.Listener
}

func (listener *ClientProxyListener) Accept() (net.Conn, error) {
	return listener.AcceptTProxy()
}

func (listener *ClientProxyListener) AcceptTProxy() (*net.TCPConn, error) {
	tcpConn, err := listener.base.(*net.TCPListener).AcceptTCP()

	if err != nil {
		return nil, err
	}
	return tcpConn, nil
	//return &ProxyConn{TCPConn: tcpConn}, nil
}

func (listener *ClientProxyListener) Addr() net.Addr {
	return listener.base.Addr()
}

func (listener *ClientProxyListener) Close() error {
	return listener.base.Close()
}

func NewClientProxyListener(network string, laddr *net.TCPAddr) (net.Listener, error) {
	//Open basic TCP listener
	listener, err := net.ListenTCP(network, laddr)
	if err != nil {
		return nil, err
	}

	//return a derived TCP listener object with TCProxy support
	return &ClientProxyListener{base: listener}, nil
}
