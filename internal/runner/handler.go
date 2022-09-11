package runner

import (
	"bufio"
	"net"
)

func sameProcessHandler() net.Conn {
	end1, end2 := net.Pipe()

	go handler(end1)

	return end2
}

func handler(c net.Conn) {
	connR := bufio.NewReader(c)
	connW := bufio.NewWriter(c)

	var b byte
	var err error

	for {
		b, err = connR.ReadByte()
		if err != nil {
			c.Close()
			return
		}

		switch MessageKind(b) {
		case BuildKind:
			err = buildCommand(connR, connW)
		case ExecuteKind:
			err = executeCommand(connR, connW)
		case ProxyKind:
			err = proxyCommand(connR, connW)
		case LogKind:
			err = logCommand(connR, connW)
		default:
			c.Close()
			return
		}

		if err != nil {
			c.Close()
			return
		}
	}
}
