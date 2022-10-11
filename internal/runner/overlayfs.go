package runner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/toastate/toastainer/internal/config"
)

// OverlayDir is a ready to mount overlaydir and tmp dir
type OverlayDir struct {
	UpperDir   string
	WorkDir    string
	MountPoint string
	uid        string
	Mounted    bool
}

// NewOverlayDir returns a ready to mount OverlayDir
// uid is the id of a btrfs subvolume
func NewOverlayDir(uid string) *OverlayDir {
	od := &OverlayDir{
		uid:        uid,
		MountPoint: filepath.Join(config.Runner.OverlayMountPoint, uid),
		WorkDir:    filepath.Join(config.Runner.BTRFSMountPoint, "overlays", uid, "workdir"),
		UpperDir:   filepath.Join(config.Runner.BTRFSMountPoint, "overlays", uid, "volume"),
	}

	err := os.MkdirAll(od.MountPoint, 0700)
	if err != nil {
		return nil
	}
	err = os.MkdirAll(od.WorkDir, 0700)
	if err != nil {
		return nil
	}
	err = os.MkdirAll(od.UpperDir, 0700)
	if err != nil {
		return nil
	}

	return od
}

// Mount an overlayfs with specified underlying directories
func (o *OverlayDir) Mount(lowerdirs []string) error {
	lower := strings.Join(lowerdirs, ":")
	opt := "lowerdir=" + lower + ",upperdir=" + o.UpperDir + ",workdir=" + o.WorkDir
	bt, err := exec.Command("/bin/mount", "-t", "overlay", "overlay", "-o", opt, o.MountPoint).CombinedOutput()
	if err != nil {
		return fmt.Errorf("Unable to mount overlaydir: %v %v", string(bt), err)
	}

	o.Mounted = true
	return nil
}

// Kill destroys an OverlayDir and its work folder
func (o *OverlayDir) Kill() {
	if o.Mounted {
		out, err := exec.Command("/bin/umount", "-f", o.MountPoint).CombinedOutput()
		if err != nil {
			fmt.Println("ERROR: could not umount overlay 1", o.uid, err, string(out))
		}
	}

	err := os.RemoveAll(o.MountPoint)
	if err != nil {
		fmt.Println("ERROR: could not remove overlay wastes 1", o.uid, err)
	}

	err = os.RemoveAll(filepath.Join(config.Runner.BTRFSMountPoint, "overlays", o.uid))
	if err != nil {
		fmt.Println("ERROR: could not remove overlay wastes 2", o.uid, err)
	}
}

func (o *OverlayDir) Umount() {
	if o.Mounted {
		out, err := exec.Command("/bin/umount", "-f", o.MountPoint).CombinedOutput()
		if err != nil {
			fmt.Println("ERROR: could not umount overlay 2", o.uid, err, string(out))
		}
		o.Mounted = false
	}
}

// Resize changes quota for an existing overlaydir
func (o *OverlayDir) Resize(sizeM int) error {
	return SetQuota(o.uid, strconv.Itoa(sizeM)+"M")
}
