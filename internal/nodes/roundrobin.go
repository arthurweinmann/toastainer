package nodes

import (
	"net"
	"sync/atomic"
)

var tvsRB uint64

func PickTVS() (t net.IP) {
	p := int(atomic.AddUint64(&tvsRB, 1))

	LocalTVSMu.RLock()
	if len(LocalTVS) > 0 {
		t = LocalTVS[p%len(LocalTVS)]
	}
	LocalTVSMu.RUnlock()

	return
}
