package utils

import (
	"fmt"
	"net"
	"syscall"
	"time"
)

//Sets additional keepalive parameters.
//Uses new interfaces introduced in Go1.11, which let us get connection's file descriptor,
//without blocking, and therefore without uncontrolled spawning of threads (not goroutines, actual threads).
func SetKeepaliveParameters(conn *net.TCPConn) error {
	conn.SetKeepAlive(true)
	conn.SetKeepAlivePeriod(time.Second * 30)

	rawConn, err := conn.SyscallConn()
	if err != nil {
		return err
	}

	return rawConn.Control(
		func(fdPtr uintptr) {
			//got socket file descriptor. Setting parameters.
			fd := int(fdPtr)
			//Number of probes.
			err := syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, syscall.TCP_KEEPCNT, 3)
			if err != nil {
				fmt.Println(err)
			}
			//Wait time after an unsuccessful probe.
			err = syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, syscall.TCP_KEEPINTVL, 3)
			if err != nil {
				fmt.Println(err)
			}
		})
}
