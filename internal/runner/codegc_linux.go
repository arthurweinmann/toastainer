package runner

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"syscall"
	"time"

	"github.com/toastate/toastcloud/internal/config"
	"github.com/toastate/toastcloud/internal/utils"
)

func runGC(gclevel int) {
	switch gclevel {
	case gcSkip:
		defer time.AfterFunc(time.Minute*10, gc)
	case gcDefault:
		defer time.AfterFunc(time.Minute*5, gc)
	case gcClean:
		defer time.AfterFunc(time.Minute*2, gc)
	case gcHard:
		defer time.AfterFunc(time.Second*15, gc)
	}

	// Calc median
	dir, err := ioutil.ReadDir(filepath.Join(config.Runner.BTRFSMountPoint, "codes"))
	if err != nil {
		utils.Warn("origin", "runGC", "warning", fmt.Sprintf("Unable to list files in %s; %v\n", filepath.Join(config.Runner.BTRFSMountPoint, "codes"), err))
		return
	}

	fileArr := make([]*codeFile, 0)

	for _, v := range dir {
		if !v.IsDir() {
			continue
		}
		fname := filepath.Join(filepath.Join(config.Runner.BTRFSMountPoint, "codes"), v.Name(), "metadata.json")
		fdir := filepath.Join(filepath.Join(config.Runner.BTRFSMountPoint, "codes"), v.Name())
		f, err := os.Stat(fname)
		if err != nil {
			log.Printf("Failed to fetch syscall.Stat_t for %s: %v\n", fname, err)
			continue
		}
		statT, ok := f.Sys().(*syscall.Stat_t)
		if !ok {
			continue
		}
		if statT == nil {
			log.Println("StatT is nil for", v.Name())
			continue
		}
		atim, _ := statT.Atim.Unix()
		fileArr = append(fileArr, &codeFile{
			fname: fdir,
			// size:    int(v.Size()),
			lastmod: atim,
		})
	}

	sort.Slice(fileArr, func(i, j int) bool { return fileArr[i].lastmod < fileArr[j].lastmod })

	var trashindex = 0
	var trashtime int64 = 0
	var trashmax = len(fileArr)
	switch gclevel {
	case gcSkip:
		trashindex = int(float64(trashmax) * 0.5)
		trashtime = time.Now().Add(time.Hour * -24 * 7).Unix()
	case gcDefault:
		trashindex = int(float64(trashmax) * 0.6)
		trashtime = time.Now().Add(time.Hour * -48).Unix()
	case gcClean:
		trashindex = int(float64(trashmax) * 0.6)
		trashtime = time.Now().Add(time.Hour * -6).Unix()
	case gcHard:
		trashindex = int(float64(trashmax) * 0.9)
		trashtime = time.Now().Add(time.Minute * -20).Unix()
	}

	trashArr := fileArr[:trashindex]
	for _, v := range trashArr {
		if v.lastmod > trashtime {
			break
		}

		err := os.RemoveAll(v.fname)
		if err != nil {
			fmt.Println("ERROR CODEGC", v.fname, err)
		}
	}
}
