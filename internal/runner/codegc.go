package runner

import (
	"fmt"
	"os"
	"time"

	"github.com/toastate/toastcloud/internal/config"
	"github.com/toastate/toastcloud/internal/utils"
)

const (
	gcSkip    = iota
	gcDefault = iota
	gcClean   = iota
	gcHard    = iota
)

var updateChtimes = make(chan string, 32) // put metadata.json paths

func initGC() {
	defer time.AfterFunc(5*time.Second, gc)
	go updateCodeAccessTimes()
}

func updateCodeAccessTimes() {
	var err error
	var p string
	var n time.Time
	for {
		p = <-updateChtimes
		n = time.Now()
		err = os.Chtimes(p, n, n)
		if err != nil {
			fmt.Println("ERROR updateCodeAccessTimes", err)
		}
	}
}

type codeFile struct {
	fname string
	// size    int
	lastmod int64
}

func gc() {
	all, free, _, err := utils.DiskUsage(config.Runner.BTRFSMountPoint)
	if err != nil {
		utils.Warn("msg", "codegc", "could not get disk usage with error", err)
		return
	}

	if all == 0 {
		utils.Warn("msg", "codegc", "could not get disk usage, running GC with Default threshold")
		runGC(gcDefault)
		return
	}

	used := 1 - (float64(free) / float64(all))

	switch {
	case used < 0.2:
		runGC(gcSkip)
	case used < 0.6:
		runGC(gcDefault)
	case used < 0.8:
		runGC(gcClean)
	default:
		runGC(gcHard)
	}
}
