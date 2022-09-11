package nodes

import (
	"fmt"
	"net"

	"github.com/toastate/toastcloud/internal/utils"
)

// To recalculate them, use gconfig/cidr/toastate_test.go::TestCalculateIPCIDRRanges
var privateCIDR = []uint32{167772160, 184549375, 2886729728, 2887778303, 3232235520, 3232301055} // 10.0.0.0/8 172.16.0.0/12 192.168.0.0/16
var loopback uint32
var localip uint32
var toasterCIDR []uint32

func initAddrs() error {
	loopback = utils.IPUint(net.ParseIP("127.0.0.1"))

	ifaces, err := net.Interfaces()
	if err != nil {
		return err
	}

	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			return err
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			ipu := utils.IPUint(ip)
			if ipu != loopback && IsIPPrivateRFC1918(ipu) {
				localip = ipu
				return nil
			}
		}
	}

	return fmt.Errorf("could not find local private net interface")
}

func IsLocalIP(ip net.IP) bool {
	ipu := utils.IPUint(ip)
	return ipu == loopback || ipu == localip
}

// IsIPPrivateRFC1918 -> See https://tools.ietf.org/html/rfc1918
func IsIPPrivateRFC1918(ipuint uint32) bool {
	if ipuint == loopback {
		return true
	}

	for i := 0; i < 6; i += 2 {
		if ipuint >= privateCIDR[i] && ipuint <= privateCIDR[i+1] {
			return true
		}
	}

	return false
}
