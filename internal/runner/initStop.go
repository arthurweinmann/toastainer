package runner

import (
	"fmt"
	"io/fs"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/shirou/gopsutil/v3/mem"
	"github.com/toastate/toastainer/internal/config"
	"github.com/toastate/toastainer/internal/db/objectstorage"
	"github.com/toastate/toastainer/internal/nodes"
	"github.com/toastate/toastainer/internal/utils"
)

var nsjailPath string
var nsjaillogs *os.File
var runnerIsShuttingDown uint32
var nodeserver *nodes.NodeServer
var maxMemoryPerToasterMega string

func Stop() {
	if atomic.CompareAndSwapUint32(&runnerIsShuttingDown, 0, 1) {
		if nodeserver != nil {
			err := nodeserver.Stop()
			if err != nil {
				utils.Error("origin", "runner.Stop", "msg", "could not stop node server", "error", err)
			}
		}
		executionWG.Wait()

		for i := 0; i < 5; i++ {
			bt, err := exec.Command("umount", "-f", config.Runner.BTRFSMountPoint).CombinedOutput()
			if err != nil && i == 4 {
				utils.Error("origin", "runner:Stop", "error", "could not umount btrfs filesystem: "+err.Error(), "details", string(bt))
				break
			}
			time.Sleep(time.Duration(i+1) * time.Second)
		}

		err := nsjaillogs.Close()
		if err != nil {
			utils.Error("origin", "runner.Stop", "msg", "could not close nsjail log file", "error", err)
		}
	}
}

// Init must be called after objectstorage initialization
func Init() error {
	var err error

	if config.Runner.BTRFSMountPoint == "" {
		config.Runner.BTRFSMountPoint = "/toastainer/btrfsmnt"
	}

	if config.Runner.OverlayMountPoint == "" {
		config.Runner.OverlayMountPoint = "/toastainer/overlaymnt"
	}

	if config.Runner.BTRFSFile == "" {
		config.Runner.BTRFSFile = "/toastainer/btrfsfile"
	}

	if config.Runner.ToasterPort == "" {
		config.Runner.ToasterPort = "8080"
	}

	// Transfer images to objectStorage if the image does not already exists in the remote location
	if !utils.DirEmpty(filepath.Join(config.Home, "images")) {
		err := filepath.Walk(filepath.Join(config.Home, "images"), func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() && strings.HasSuffix(path, ".tar.gz") && len(filepath.Base(path)) > len(".tar.gz") {
				imgName := filepath.Base(path)[:len(filepath.Base(path))-len(".tar.gz")]
				exists, err := objectstorage.Client.Exists(filepath.Join("images", imgName))
				if err != nil {
					return fmt.Errorf("objectstorage error: %v", err)
				}
				if !exists {
					f, err := os.Open(path)
					if err != nil {
						return err
					}
					defer f.Close()

					err = objectstorage.Client.PushReader(f, filepath.Join("images", imgName))
					if err != nil {
						return err
					}
				}
			}

			return nil
		})
		if err != nil {
			return err
		}
	}

	vm, err := mem.VirtualMemory()
	if err != nil {
		return fmt.Errorf("could not evalutate available virtual memory")
	}

	maxMemoryPerToasterMega = strconv.Itoa(int(utils.Min((vm.Available / 25 * 1024 * 1024), 256)))

	nsjaillogs, err = os.OpenFile(filepath.Join(config.Home, "nsjail.log"), os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}

	nsjailPath = filepath.Join(config.Home, "nsjail")

	incrQueue = make(chan string, 512)
	decrQueue = make(chan *DecrReq, 512)
	startRedisQueueRoutine()

	err = buildDisk()
	if err != nil {
		return err
	}
	btrfsPoolFillRoutine()

	if config.NodeDiscovery {
		nodeserver, err = nodes.StartNodeServer(net.ParseIP(config.LocalPrivateIP).To4(), handler)
		if err != nil {
			utils.Error("origin", "nodes.StartNodeServer", "error", err)
		}
	}

	initGC()

	return nil
}

func buildDisk() error {
	err := os.MkdirAll(config.Runner.BTRFSMountPoint, 0700)
	if err != nil {
		return err
	}

	err = os.Chown(config.Runner.BTRFSMountPoint, config.Runner.NonRootUID, config.Runner.NonRootGID)
	if err != nil {
		return err
	}

	err = os.MkdirAll(config.Runner.OverlayMountPoint, 0755)
	if err != nil {
		return err
	}

	err = os.Chown(config.Runner.OverlayMountPoint, config.Runner.NonRootUID, config.Runner.NonRootGID)
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
			return fmt.Errorf("it seems a btrfs filesystem is already mounted but %s is not present", activeFile)
		}

		stats, err := GetBTRFSStats()
		if err != nil {
			return err
		}

		if len(stats) > 0 && (len(stats) > 1 || (stats[0].Devices != nil && len(stats[0].Devices) > 0)) {
			devices := ""
			for i := 0; i < len(stats); i++ {
				for n := range stats[i].Devices {
					devices += n + ","
				}
			}
			utils.Warn("msg", fmt.Sprintf("WARNING: it seems there are existing not mounted phantom btrfs file system device: %v\n", devices))
		}

		if config.Runner.UseUnmountedDisks {
			utils.Info("origin", "runner.Init", "msg", "mounting and formating all unmounted disks")

			umounted = listUnmounted()

			if len(umounted) > 0 {
				primDisk = "/dev/" + umounted[0]
				umounted = umounted[1:]
			} else {
				utils.Info("origin", "runner.Init", "msg", "no unmounted disk found")
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

			utils.Info("origin", "runner.Init", "msg", fmt.Sprintf("Mounting formated file %s as file system for Toasters with %v bytes", primDisk, config.Runner.BTRSFileSize))

			_, err = os.Stat(primDisk)
			if err != nil && !os.IsNotExist(err) {
				return err
			}

			if err != nil {
				err = os.MkdirAll(filepath.Dir(primDisk), 0700)
				if err != nil {
					return err
				}

				f, err := os.OpenFile(primDisk, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0600)
				if err != nil {
					return err
				}

				err = f.Truncate(config.Runner.BTRSFileSize)
				f.Close()
				if err != nil {
					return err
				}
			}
		}

		utils.Info("origin", "runner.Init", "msg", "btrfs formatting primDisk", "identifier", primDisk)

		bt, err := formatBtrfs(primDisk)
		if err != nil {
			return fmt.Errorf("diskutils error: format partition: %v %v", string(bt), err)
		}

		bt, err = exec.Command("mount", primDisk, config.Runner.BTRFSMountPoint).CombinedOutput()
		if err != nil {
			return fmt.Errorf("diskutils error: mount partition %v %v", string(bt), err)
		}

		f, err := os.Create(activeFile)
		if err != nil {
			return fmt.Errorf("diskutils error: active %v", err)
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
			return err
		}

		err = os.Chown(p, config.Runner.NonRootUID, config.Runner.NonRootGID)
		if err != nil {
			return err
		}
	}

	if !utils.DirExists(filepath.Join(config.Runner.BTRFSMountPoint, "codes")) {
		p, err := NewSubvolume("codes", "")
		if err != nil {
			return err
		}

		err = os.Chown(p, config.Runner.NonRootUID, config.Runner.NonRootGID)
		if err != nil {
			return err
		}
	}

	if !utils.DirExists(filepath.Join(config.Runner.BTRFSMountPoint, "exe")) {
		p, err := NewSubvolume("exe", "")
		if err != nil {
			return err
		}

		err = os.Chown(p, config.Runner.NonRootUID, config.Runner.NonRootGID)
		if err != nil {
			return err
		}
	}

	if !utils.DirExists(filepath.Join(config.Runner.BTRFSMountPoint, "overlays")) {
		p, err := NewSubvolume("overlays", "")
		if err != nil {
			return err
		}

		err = os.Chown(p, config.Runner.NonRootUID, config.Runner.NonRootGID)
		if err != nil {
			return err
		}
	}

	if len(umounted) > 0 && config.Runner.UseUnmountedDisks {
		go func() {
			for _, v := range umounted {
				utils.Info("origin", "runner.Init", "msg", "diskutils: found disk /dev/"+v+" ,adding")
				cmd := exec.Command("btrfs", "device", "add", "-f", "/dev/"+v, config.Runner.BTRFSMountPoint)

				bt, err := cmd.CombinedOutput()
				if err != nil {
					utils.Error("origin", "runner.Init:umountedDisksBackgroundRoutine", "error", fmt.Sprintf("diskutils error: add disk %v: %v: %v", cmd.Args, err, string(bt)))
					return
				}
			}
			utils.Info("origin", "runner.Init", "msg", "diskutils: all umounted disks mounted")
		}()
	}

	return nil
}
