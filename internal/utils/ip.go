package utils

import (
	"fmt"
	"net"
)

var LoopbackIPUint uint32
var PrivateCIDRUint = []uint32{167772160, 184549375, 2886729728, 2887778303, 3232235520, 3232301055} // 10.0.0.0/8 172.16.0.0/12 192.168.0.0/16

func init() {
	LoopbackIPUint = IPUint(net.ParseIP("127.0.0.1"))
}

func IPUint(ip net.IP) uint32 {
	if ip == nil {
		return 0
	}
	ip = ip.To4()
	if ip == nil {
		return 0
	}

	return uint32(ip[3]) | uint32(ip[2])<<8 | uint32(ip[1])<<16 | uint32(ip[0])<<24
}

func IPUintSl(ip []byte) uint32 {
	return uint32(ip[3]) | uint32(ip[2])<<8 | uint32(ip[1])<<16 | uint32(ip[0])<<24
}

func UintIP(v uint32) net.IP {
	v3 := byte(v & 0xFF)
	v2 := byte((v >> 8) & 0xFF)
	v1 := byte((v >> 16) & 0xFF)
	v0 := byte((v >> 24) & 0xFF)
	return net.IPv4(v0, v1, v2, v3).To4()
}

func IncrIP(ip net.IP, inc uint) net.IP {
	i := ip.To4()
	v := uint(i[0])<<24 + uint(i[1])<<16 + uint(i[2])<<8 + uint(i[3])
	v += inc
	v3 := byte(v & 0xFF)
	v2 := byte((v >> 8) & 0xFF)
	v1 := byte((v >> 16) & 0xFF)
	v0 := byte((v >> 24) & 0xFF)
	return net.IPv4(v0, v1, v2, v3).To4()
}

func DecrIP(ip net.IP, inc uint) net.IP {
	i := ip.To4()
	v := uint(i[0])<<24 + uint(i[1])<<16 + uint(i[2])<<8 + uint(i[3])
	v -= inc
	v3 := byte(v & 0xFF)
	v2 := byte((v >> 8) & 0xFF)
	v1 := byte((v >> 16) & 0xFF)
	v0 := byte((v >> 24) & 0xFF)
	return net.IPv4(v0, v1, v2, v3).To4()
}

func GetLocalPrivateIP() (net.IP, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			return nil, err
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			ipu := IPUint(ip)
			if ipu != LoopbackIPUint && IsIPPrivateRFC1918(ipu) {
				return ip, nil
			}
		}
	}

	return nil, fmt.Errorf("could not find local private net interface")
}

func GetLocalPublicIP() (net.IP, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			return nil, err
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			ipu := IPUint(ip)
			if ipu != LoopbackIPUint && !IsIPPrivateRFC1918(ipu) {
				return ip, nil
			}
		}
	}

	return nil, fmt.Errorf("could not find local private net interface")
}

// IsIPPrivateRFC1918 -> See https://tools.ietf.org/html/rfc1918
func IsIPPrivateRFC1918(ipuint uint32) bool {
	if ipuint == LoopbackIPUint {
		return true
	}

	for i := 0; i < 6; i += 2 {
		if ipuint >= PrivateCIDRUint[i] && ipuint <= PrivateCIDRUint[i+1] {
			return true
		}
	}

	return false
}
