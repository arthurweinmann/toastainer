package runner

import (
	"errors"
	"fmt"
	"io/fs"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/prometheus/procfs/btrfs"
	"github.com/rs/xid"
	"github.com/toastate/toastainer/internal/config"
)

const (
	maxPooledVolumes = 64
)

var btrfsPool = make(chan string, maxPooledVolumes)
var btrfsDelPool = make(chan string, maxPooledVolumes*2)

func btrfsPoolFillRoutine() {
	// Clean subvolumes from pool from previous running tvs that shut down impromptly
	err := filepath.Walk(filepath.Join(config.Runner.BTRFSMountPoint, "exe"), func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && path != filepath.Join(config.Runner.BTRFSMountPoint, "exe") {
			err = DelSubvolumeAbsolute(path)
			if err != nil {
				return fmt.Errorf("%s: %v", path, err)
			}
		}

		return nil
	})
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			// only for executions and compilations
			p, err := NewSubvolume(filepath.Join("exe", xid.New().String()), "200M")
			if err != nil {
				fmt.Println("BTRFS Fill Routine", err)
				time.Sleep(2 * time.Second)
				continue
			}

			btrfsPool <- p

			select {
			case p = <-btrfsDelPool:
				err = DelSubvolumeAbsolute(p)
				if err != nil {
					fmt.Println("BTRFS Fill Routine, delete", p, ", ", err)
					continue
				}
			default:
			}
		}
	}()
}

// TODO: split this command in 2 to also detect unmounted disk already partitioned as btrfs we only need to mount and not to format
const findUnmountedDisks = `lsblk | \
    grep disk | \
    grep -Eo ^[a-zA-Z0-9]+ | \
    xargs -I % sh -c 'mount | grep -o % > /dev/null 2>&1 || echo %'| \
	xargs -I % sh -c 'btrfs filesystem show | grep -o % > /dev/null 2>&1 || echo %'`

func listUnmounted() []string {
	ret, err := exec.Command("bash", "-c", findUnmountedDisks).CombinedOutput()
	if err != nil {
		return []string{}
	}
	return strings.Fields(string(ret))
}

func GetBTRFSStats() ([]*btrfs.Stats, error) {
	fs, err := btrfs.NewDefaultFS()
	if err != nil {
		return nil, err
	}

	return fs.Stats()
}

func IsBTRFSMounted(folder string) bool {
	ret, err := exec.Command("/bin/bash", "-c", `mount | grep "`+folder+`" | grep "type btrfs" && echo done`).CombinedOutput()
	if err != nil {
		return false
	}
	if strings.Index(string(ret), "done") < 0 {
		return false
	}

	_, err = exec.Command("/bin/bash", "-c", `btrfs filesystem show `+folder).CombinedOutput()
	if err != nil {
		return false
	}

	return true
}

// NewSubvolume creates new btrfs subvolume with given quota
func NewSubvolume(name string, size string) (string, error) {
	p := filepath.Join(config.Runner.BTRFSMountPoint, name)
	bt, err := exec.Command("/bin/btrfs", "subvolume", "create", p).CombinedOutput()
	if err != nil {
		stats, err := GetBTRFSStats()
		if err != nil {
			panic(err)
		}
		spew.Dump(stats)

		fmt.Println("diskutils error: create subvolume", name, size, string(bt), err)
		return "", errors.New("diskutils error: create subvolume")
	}

	if size != "" {
		bt, err = exec.Command("/bin/btrfs", "qgroup", "limit", size, p).CombinedOutput()
		if err != nil {
			stats, err := GetBTRFSStats()
			if err != nil {
				panic(err)
			}
			spew.Dump(stats)

			fmt.Println("diskutils error: create subvolume qgroup limit", name, size, string(bt), err)
			return "", errors.New("diskutils error: create subvolume qgroup limit")
		}
	}
	return p, nil
}

// NewSubvolumeAbsolute does the same as NewSubvolume but with a subvolume absolute path
func NewSubvolumeAbsolute(p string, size string) (string, error) {
	bt, err := exec.Command("/bin/btrfs", "subvolume", "create", p).CombinedOutput()
	if err != nil {
		fmt.Println("diskutils error: create subvolume", p, size, string(bt), err)
		return "", errors.New("diskutils error: create subvolume")
	}

	if size != "" {
		exec.Command("/bin/btrfs", "qgroup", "limit", size, p).CombinedOutput()
		if err != nil {
			fmt.Println("diskutils error: create subvolume", p, size, string(bt), err)
			return "", errors.New("diskutils error: create subvolume")
		}
	}
	return p, nil
}

func DelSubvolume(name string) error {
	p := filepath.Join(config.Runner.BTRFSMountPoint, name)
	bt, err := exec.Command("/bin/btrfs", "subvolume", "delete", p).CombinedOutput()
	if err != nil {
		return fmt.Errorf("DelSubvolume %v: %v", err, string(bt))
	}

	return nil
}

func DelSubvolumeAbsolute(p string) error {
	bt, err := exec.Command("/bin/btrfs", "subvolume", "delete", p).CombinedOutput()
	if err != nil {
		return fmt.Errorf("DelSubvolumeAbsolute %v: %v", err, string(bt))
	}

	return nil
}

func Snapshot(src string, size string) (string, error) {
	uid := "snap_" + xid.New().String()
	p := filepath.Join(config.Runner.BTRFSMountPoint, uid)

	bt, err := exec.Command("/bin/btrfs", "subvolume", "snapshot", src, p).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("diskutils error: snapshot subvolume %v %v %v %v", src, size, string(bt), err)
	}

	// fmt.Println("snapshot output", string(bt))

	if size != "" {
		err = SetQuota(uid, size)
		if err != nil {
			return "", err
		}
	}

	return p, nil
}

func SnapshotAbsolute(src, dest string, size string) error {
	bt, err := exec.Command("/bin/btrfs", "subvolume", "snapshot", src, dest).CombinedOutput()
	if err != nil {
		return fmt.Errorf("diskutils error: snapshot subvolume %v %v %v %v", src, size, string(bt), err)
	}

	// fmt.Println("snapshot output", string(bt))

	if size != "" {
		err = SetQuota(dest, size)
		if err != nil {
			return err
		}
	}

	return nil
}

// SetQuota changes btrfs subvolume quota
func SetQuota(subvolume string, size string) error {
	bt, err := exec.Command("/bin/btrfs", "qgroup", "limit", size, filepath.Join(config.Runner.BTRFSMountPoint, subvolume)).CombinedOutput()
	if err != nil {
		return fmt.Errorf("diskutils error: set subvolume quota: %v %v", err, string(bt))
	}
	return nil
}

func primaryPartition(diskPath string) string {
	devEntry := filepath.Base(diskPath)
	devPath := filepath.Dir(diskPath)
	if len(devEntry) >= 4 && devEntry[:4] == "nvme" {
		return filepath.Join(devPath, devEntry+"p1")
	}
	return filepath.Join(devPath, devEntry+"1")
}

func formatBtrfs(partition string) ([]byte, error) {
	/*
	* Use full capacity of multiple drives with different sizes (metadata mirrored, data not mirrored and not striped)
	* TODO: test if we should put metadata on a single disk and not replicated too with -m single
	* See: https://btrfs.wiki.kernel.org/index.php/Using_Btrfs_with_Multiple_Devices ; https://www.mail-archive.com/linux-btrfs@vger.kernel.org/msg44523.html
	* and https://lwn.net/Articles/577961/
	 */
	return exec.Command("/bin/bash", "-c", "mkfs.btrfs -f -d single "+partition).CombinedOutput()
}
