package config

import (
	"fmt"
	"net"

	"github.com/toastate/toastainer/internal/utils"
)

var LocalPrivIPUint64 uint32
var LocalPubIPUint64 uint32

func initIPconf() error {
	if LocalPrivateIP != "" {
		tmp := net.ParseIP(LocalPrivateIP)
		if tmp == nil {
			return fmt.Errorf("invalid local private IPV4 %s in configuration file", LocalPrivateIP)
		}

		tmp = tmp.To4()
		if tmp == nil {
			return fmt.Errorf("invalid local private IPV4 %s in configuration file", LocalPrivateIP)
		}

		ipu := utils.IPUint(tmp)
		if ipu == utils.LoopbackIPUint || !utils.IsIPPrivateRFC1918(ipu) {
			return fmt.Errorf("invalid local private IPV4 %s in configuration file", LocalPrivateIP)
		}

		LocalPrivIPUint64 = ipu
		LocalPrivateIP = tmp.String()
	} else {
		tmp, err := utils.GetLocalPrivateIP()
		if err != nil {
			return err
		}
		LocalPrivIPUint64 = utils.IPUint(tmp)
		LocalPrivateIP = tmp.String()
	}

	tmp, err := utils.GetLocalPublicIP()
	if err != nil {
		return fmt.Errorf("could not find local public IP address: %v", err)
	}
	LocalPubIPUint64 = utils.IPUint(tmp)
	LocalPublicIP = tmp.String()

	return nil
}

func IsLocalIP(ip net.IP) bool {
	ipu := utils.IPUint(ip)
	return ipu == utils.LoopbackIPUint || ipu == LocalPubIPUint64 || ipu == LocalPrivIPUint64
}
