package nodes

import (
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hashicorp/yamux"
	"github.com/toastate/toastcloud/internal/utils"
)

const (
	connpoolShards = 32
	maxConnPerPool = 16
	port           = "5555"
	portInt        = 5555
)

type pool struct {
	conns     []*yamux.Session // must be secured through wireguard
	connsMu   sync.RWMutex
	roudRobin uint64
}

var pools []map[uint32]*pool
var poolsMu []*sync.RWMutex

func init() {
	for i := 0; i < connpoolShards; i++ {
		pools = append(pools, map[uint32]*pool{})
		poolsMu = append(poolsMu, &sync.RWMutex{})
	}
}

func newPool() *pool {
	return &pool{}
}

// GetConn add.pool must not be nil
func GetConn(ip net.IP) (net.Conn, error) {
	ipuint := utils.IPUint(ip)
	if !IsIPPrivateRFC1918(ipuint) {
		return nil, fmt.Errorf("IP is not a private one")
	}

	shard := ipuint % connpoolShards

	poolsMu[shard].RLock()
	p, ok := pools[shard][ipuint]
	if !ok {
		p = newPool()
		pools[shard][ipuint] = p
	}
	poolsMu[shard].RUnlock()

	if !ok {
		err := p.fill(ip)
		if err != nil {
			return nil, err
		}
	}

	rrb := int(atomic.AddUint64(&p.roudRobin, 1))
	p.connsMu.RLock()
	session := p.conns[rrb%len(p.conns)]
	p.connsMu.RUnlock()

	stream, err := session.Open()
	if err != nil {
		return nil, err
	}

	return stream, nil
}

func PutConn(ip net.IP, conn net.Conn) {
	conn.Close()
}

var dialer = net.Dialer{Timeout: 10 * time.Second}

func Dial(ip net.IP) (*yamux.Session, error) {
	netconn, err := dialer.Dial("tcp4", ip.String()+":"+port)
	if err != nil {
		return nil, err
	}

	conn := netconn.(*net.TCPConn)

	err = utils.SetKeepaliveParameters(conn)
	if err != nil {
		return nil, err
	}

	return yamux.Client(conn, nil)
}

func (p *pool) fill(ip net.IP) error {
	first, err := Dial(ip)
	if err != nil {
		return err
	}

	p.conns = append(p.conns, first)

	go func() {
		for i := 1; i < maxConnPerPool; i++ {
			sess, err := Dial(ip)
			if err != nil {
				fmt.Println("pool fill", err)
				time.Sleep(60 * time.Second)
				continue
			}

			p.connsMu.Lock()
			p.conns = append(p.conns, sess)
			p.connsMu.Unlock()

			time.Sleep(30 * time.Second)
		}
	}()

	return nil
}
