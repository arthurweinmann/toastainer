package smartdhcp

const dhcpPoolOffset = 257
const dhcpPoolSize = 512

var dhcpPool = make(chan int64, dhcpPoolSize)

func init() {
	var i int64
	for i = dhcpPoolOffset; i < dhcpPoolSize+dhcpPoolOffset; i++ {
		end := i & 0xFF
		beg := (i &^ end) >> 8
		if end == 0 || end == 255 {
			continue
		}
		if beg == 0 || beg == 255 {
			continue
		}
		dhcpPool <- i
	}
}

// Get returns a usable IP address
func Get() int64 {
	return <-dhcpPool
}

// Put releases an IP address
func Put(i int64) {
	dhcpPool <- i
}
