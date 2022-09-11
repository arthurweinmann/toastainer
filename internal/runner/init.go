package runner

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/toastate/toastcloud/internal/config"
	"github.com/toastate/toastcloud/internal/nodes"
	"github.com/toastate/toastcloud/internal/utils"
)

var DebugMode bool
var nsjailPAth string

func WaitAllExeEnd() {
	executionWG.Wait()
}

func Init() error {
	nsjailPAth = filepath.Join(config.Home, "nsjail")

	incrQueue = make(chan string, 512)
	decrQueue = make(chan *DecrReq, 512)
	startRedisQueueRoutine()

	err := buildDisk()
	if err != nil {
		return err
	}
	btrfsPoolFillRoutine()

	if config.NodeDiscovery {
		// Use Connect2Any and Connect2 functions to connect to a runner because they will handle the edge case of a runner
		// being in the same process

		go func() {
			err = nodes.StartNodeServer(net.ParseIP(config.LocalPrivateIP).To4(), handler)
			if err != nil {
				fmt.Println("nodes.StartNodeServer", err)
			}
		}()
	}

	initGC()

	return nil
}

func buildDisk() error {
	err := os.MkdirAll(config.Runner.BTRFSMountPoint, 0700)
	if err != nil {
		return err
	}

	err = os.MkdirAll(config.Runner.OverlayMountPoint, 0755)
	if err != nil {
		return err
	}

	var primDisk = ""
	var umounted []string
	activeFile := filepath.Join(config.Runner.BTRFSMountPoint, "active")

	_, err = os.Stat(activeFile)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}

		// Maybe use btrfs device scan command to find already existing btrfs filesystem ?
		if IsBTRFSMounted(config.Runner.BTRFSMountPoint) {
			panic(fmt.Errorf("It seems a btrfs filesystem is already mounted but %s is not present", activeFile))
		}

		stats, err := GetBTRFSStats()
		if err != nil {
			return err
		}

		if stats != nil && len(stats) > 0 && (len(stats) > 1 || (stats[0].Devices != nil && len(stats[0].Devices) > 0)) {
			devices := ""
			for i := 0; i < len(stats); i++ {
				for n := range stats[i].Devices {
					devices += n + ","
				}
			}
			fmt.Printf("WARNING: it seems there are existing not mounted phantom btrfs file system device: %v\n", devices)
		}

		if config.Runner.UseUnmountedDisks {
			fmt.Println("mounting and formating all unmounted disks")

			umounted = listUnmounted()

			if umounted != nil && len(umounted) > 0 {
				primDisk = "/dev/" + umounted[0]
				umounted = umounted[1:]
			} else {
				fmt.Println("no unmounted disk found")
			}
		}

		if primDisk == "" {
			primDisk = config.Runner.BTRFSFile

			if config.Runner.BTRSFileSize == 0 {
				_, free, _, err := utils.DiskUsage("/")
				if err != nil {
					return err
				}
				config.Runner.BTRSFileSize = int64(float64(free) * 0.5)
			}

			fmt.Println("Mounting formated file", primDisk, "as file system for Toasters with", config.Runner.BTRSFileSize, "bytes")

			_, err = os.Stat(primDisk)
			if err != nil && !os.IsNotExist(err) {
				return err
			}

			if err != nil {
				err = os.MkdirAll(filepath.Dir(primDisk), 0700)
				if err != nil {
					panic(err)
				}
				f, err := os.OpenFile(primDisk, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 600) // user has read/write perms, group and all others have no permissions
				if err != nil {
					panic(err)
				}

				err = f.Truncate(config.Runner.BTRSFileSize)
				if err != nil {
					f.Close()
					panic(err)
				}
				f.Close()
			}
		}

		fmt.Println("btrfs formatting primDisk", primDisk)

		bt, err := formatBtrfs(primDisk)
		if err != nil {
			return fmt.Errorf("diskutils error: format partition: %v %v", string(bt), err)
		}

		// mount it
		bt, err = exec.Command("mount", primDisk, config.Runner.BTRFSMountPoint).CombinedOutput()
		if err != nil {
			return fmt.Errorf("diskutils error: mount partition %v %v", string(bt), err)
		}

		f, err := os.Create(activeFile)
		if err != nil {
			panic(fmt.Errorf("diskutils error: active %v", err))
		}
		f.Close()
	}

	bt, err := exec.Command("btrfs", "quota", "enable", config.Runner.BTRFSMountPoint).CombinedOutput()
	if err != nil {
		return fmt.Errorf("diskutils error: quota %v %v", string(bt), err)
	}

	if !utils.DirExists(filepath.Join(config.Runner.BTRFSMountPoint, "images")) {
		p, err := NewSubvolume("images", "")
		if err != nil {
			panic(err)
		}

		err = os.Chown(p, config.Runner.NonRootUID, config.Runner.NonRootGID)
		if err != nil {
			panic(err)
		}
	}

	if !utils.DirExists(filepath.Join(config.Runner.BTRFSMountPoint, "codes")) {
		p, err := NewSubvolume("codes", "")
		if err != nil {
			panic(err)
		}

		err = os.Chown(p, config.Runner.NonRootUID, config.Runner.NonRootGID)
		if err != nil {
			panic(err)
		}
	}

	if !utils.DirExists(filepath.Join(config.Runner.BTRFSMountPoint, "exe")) {
		p, err := NewSubvolume("exe", "")
		if err != nil {
			panic(err)
		}

		err = os.Chown(p, config.Runner.NonRootUID, config.Runner.NonRootGID)
		if err != nil {
			panic(err)
		}
	}

	if !utils.DirExists(filepath.Join(config.Runner.BTRFSMountPoint, "overlays")) {
		p, err := NewSubvolume("overlays", "")
		if err != nil {
			panic(err)
		}

		err = os.Chown(p, config.Runner.NonRootUID, config.Runner.NonRootGID)
		if err != nil {
			panic(err)
		}
	}

	if umounted != nil && len(umounted) > 0 && config.Runner.UseUnmountedDisks {
		go func() {
			for _, v := range umounted {
				fmt.Println("diskutils: found disk /dev/" + v + " ,adding")
				cmd := exec.Command("btrfs", "device", "add", "-f", "/dev/"+v, config.Runner.BTRFSMountPoint)

				bt, err := cmd.CombinedOutput()
				if err != nil {
					fmt.Println("ERROR: diskutils error: add disk", cmd.Args, string(bt), err)
					return
				}
			}
			fmt.Println("diskutils: all umounted disks mounted")
		}()
	}

	return nil
}
