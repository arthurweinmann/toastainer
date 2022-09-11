package nodes

import (
	"fmt"
	"net"
	"strings"

	"github.com/hashicorp/yamux"
	"github.com/toastate/toastcloud/internal/utils"
)

const (
	// initialStreamWindow is the initial stream window size
	initialStreamWindow uint32 = 256 * 1024
)

var (
	toasterIPSpace = []uint32{utils.IPUint(net.ParseIP("10.166.0.0")), utils.IPUint(net.ParseIP("10.166.255.255"))}
)

var list *net.TCPListener

// TODO: generalize StartNodeServer to start a server for each node that shares runner comm and discovery auth and registration in addition of dns discovery
func StartNodeServer(listenToIP net.IP, handler func(net.Conn)) error {
	var err error

	list, err = net.ListenTCP("tcp4", &net.TCPAddr{IP: listenToIP, Port: portInt})
	if err != nil {
		return err
	}

	for {
		conn, err := list.AcceptTCP()
		if err != nil {
			panic(err)
		}

		remoteipuint := utils.IPUint(net.ParseIP(strings.Split(conn.RemoteAddr().String(), ":")[0]))
		if !IsIPPrivateRFC1918(remoteipuint) ||
			!IsIPPrivateRFC1918(utils.IPUint(net.ParseIP(strings.Split(conn.LocalAddr().String(), ":")[0]))) ||
			(remoteipuint >= toasterIPSpace[0] && remoteipuint <= toasterIPSpace[1]) {
			fmt.Println("Node server: received connection request from non private remote or to non private local address:", conn.RemoteAddr().String(), conn.LocalAddr().String())
			conn.Close()
			continue
		}

		err = utils.SetKeepaliveParameters(conn)
		if err != nil {
			fmt.Println("Node server:", err)
			conn.Close()
			continue
		}

		session, err := yamux.Server(conn, nil)
		if err != nil {
			fmt.Println("Node server:", err)
			conn.Close()
			continue
		}

		go func(session *yamux.Session, conn *net.TCPConn) {
			for {
				stream, err := session.Accept()
				if err != nil {
					fmt.Println("Node server:", err)
					conn.Close()
					return
				}

				go handler(stream)
			}
		}(session, conn)
	}
}
