package runner

import (
	"fmt"
	"net"
	"sync/atomic"
	"time"

	"github.com/toastate/toastainer/internal/config"
	"github.com/toastate/toastainer/internal/nodes"
)

type internalPipe struct {
	net.Conn
}

type tcpConn struct {
	net.Conn
	runnerip net.IP
	closed   uint32
}

// Use Connect2Any and Connect2 functions to connect to a runner because they will handle wether the runner is local and in
// the same process or on a remote server
func Connect2Any() (net.Conn, error) {
	if config.NodeDiscovery {
		var err error
		var conn net.Conn
		for i := 0; i < 3; i++ {
			runnerip := nodes.PickTVS()
			if config.IsLocalIP(runnerip) {
				return &internalPipe{sameProcessHandler()}, nil
			} else {
				conn, err = nodes.GetConn(runnerip)
				if err == nil {
					return &tcpConn{conn, runnerip, 0}, nil
				}
				if conn != nil {
					conn.Close()
				}
			}

			time.Sleep(300 * time.Millisecond)
		}

		return nil, fmt.Errorf("could not handshake a build server: %v", err)
	}

	return &internalPipe{sameProcessHandler()}, nil
}

// Use Connect2Any and Connect2 functions to connect to a runner because they will handle wether the runner is local and in
// the same process or on a remote server
func Connect2(runnerip net.IP) (net.Conn, error) {
	if config.NodeDiscovery {
		if config.IsLocalIP(runnerip) {
			return &internalPipe{sameProcessHandler()}, nil
		} else {
			conn, err := nodes.GetConn(runnerip)
			if err != nil {
				if conn != nil {
					conn.Close()
				}
				return nil, fmt.Errorf("could not handshake a build server: %v", err)
			}

			return &tcpConn{conn, runnerip, 0}, nil
		}
	}

	return &internalPipe{sameProcessHandler()}, nil
}

func PutConnection(c net.Conn) {
	switch t := c.(type) {
	case *internalPipe:
		t.Conn.Close()
	case *tcpConn:
		if atomic.CompareAndSwapUint32(&t.closed, 0, 1) {
			nodes.PutConn(t.runnerip, t.Conn)
		}
	default:
		panic(fmt.Errorf("%T", c))
	}
}

func (conn *tcpConn) Close() error {
	if atomic.CompareAndSwapUint32(&conn.closed, 0, 1) {
		return conn.Conn.Close()
	}

	return nil
}

func (conn *tcpConn) Read(b []byte) (n int, err error) {
	return conn.Conn.Read(b)
}

func (conn *tcpConn) Write(b []byte) (n int, err error) {
	return conn.Conn.Write(b)
}

func (conn *tcpConn) LocalAddr() net.Addr {
	return conn.Conn.LocalAddr()
}

func (conn *tcpConn) RemoteAddr() net.Addr {
	return conn.Conn.RemoteAddr()
}

func (conn *tcpConn) SetDeadline(t time.Time) error {
	return conn.Conn.SetDeadline(t)
}

func (conn *tcpConn) SetReadDeadline(t time.Time) error {
	return conn.Conn.SetReadDeadline(t)
}

func (conn *tcpConn) SetWriteDeadline(t time.Time) error {
	return conn.Conn.SetWriteDeadline(t)
}
