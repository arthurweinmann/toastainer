package nodes

import (
	"github.com/toastate/toastainer/internal/config"
)

func Init() error {
	if !config.NodeDiscovery {
		return nil
	}

	startDNSNodeLookupRoutine()

	return nil
}
