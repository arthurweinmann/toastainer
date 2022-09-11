package utils

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var GoogleResolver = &net.Resolver{
	PreferGo: true,
	Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
		d := net.Dialer{
			Timeout: time.Millisecond * time.Duration(10000),
		}
		return d.DialContext(ctx, network, "8.8.8.8:53")
	},
}

var R interface {
	LookupAddr(ctx context.Context, addr string) ([]string, error)
	LookupCNAME(ctx context.Context, host string) (string, error)
	LookupHost(ctx context.Context, host string) (addrs []string, err error)
	// LookupIP(ctx context.Context, network, host string) ([]net.IP, error)
	LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error)
	LookupMX(ctx context.Context, name string) ([]*net.MX, error)
	LookupNS(ctx context.Context, name string) ([]*net.NS, error)
	LookupPort(ctx context.Context, network, service string) (port int, err error)
	LookupSRV(ctx context.Context, service, proto, name string) (string, []*net.SRV, error)
	LookupTXT(ctx context.Context, name string) ([]string, error)
} = GoogleResolver

func ForceIPHTTPClient(ip, port string) *http.Client {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		// DualStack: true, // this is deprecated as of go 1.16
	}

	return &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return dialer.DialContext(ctx, network, ip+":"+port)
			},
		},
		Timeout: 60 * time.Second,
	}
}

func ForceIPWebsocketDialer(ip, port string) *websocket.Dialer {
	netdialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		// DualStack: true, // this is deprecated as of go 1.16
	}

	return &websocket.Dialer{
		NetDial: func(network, addr string) (net.Conn, error) {
			return netdialer.DialContext(context.Background(), network, ip+":"+port)
		},
		NetDialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return netdialer.DialContext(ctx, network, ip+":"+port)
		},
		HandshakeTimeout: 60 * time.Second,
	}
}
