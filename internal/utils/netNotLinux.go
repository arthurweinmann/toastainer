//go:build !linux

package utils

import "net"

func SetKeepaliveParameters(conn *net.TCPConn) error {
	panic("SetKeepaliveParameters only available on linux systems")
}
