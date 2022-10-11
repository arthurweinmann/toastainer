package runner

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/toastate/toastcloud/internal/config"
	"github.com/toastate/toastcloud/internal/db/objectstorage"
	"github.com/toastate/toastcloud/internal/runner/smartdhcp"
	"github.com/toastate/toastcloud/internal/utils"
)

// Env golang compile: "GOPATH=/home/ubuntu/go", "GOROOT=/usr/local/go", "TERM=xterm-color", "HOME=/home/ubuntu", "PATH=/home/ubuntu/go/bin:/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"

// source code remain in the execution image
// you can manually delete it in build command
type BuildCommand struct {
	CodeID   string
	Image    string
	BuildCmd []string
	Env      []string
}

func buildCommand(connR *bufio.Reader, connW *bufio.Writer) (err error) {
	defer func() {
		if err != nil {
			err2 := writeError(connW, err)
			if err2 != nil {
				utils.Error("origin", "runner:buildCommand", "error", fmt.Sprintf("could not write error: %v", err2))
			}
		}
	}()

	var b []byte
	b, err = readCommand(connR)
	if err != nil {
		return err
	}

	cmd := &BuildCommand{}
	err = json.Unmarshal(b, cmd)
	if err != nil {
		return
	}

	var logs []byte
	logs, err = buildCommandInternal(cmd, connR)
	if err != nil {
		return
	}

	err = writeSuccessRaw(connW, logs)
	return
}

func buildCommandInternal(cmd *BuildCommand, connR *bufio.Reader) (logs []byte, err error) {
	var pimg string
	pimg, err = pullImg(cmd.Image)
	if err != nil {
		return nil, fmt.Errorf("pullImg %s: %v", cmd.Image, err)
	}

	var pcompil string
	select {
	case pcompil = <-btrfsPool:
	case <-time.After(10 * time.Second):
		return nil, fmt.Errorf("could not create compilation volume")
	}
	defer func() {
		select {
		case btrfsDelPool <- pcompil:
		default:
			err := DelSubvolumeAbsolute(pcompil)
			if err != nil {
				fmt.Println("could not delete subvolume", err)
			}
		}
	}()

	ovlr := NewOverlayDir(filepath.Base(pcompil))

	err = ovlr.Mount([]string{pimg})
	if err != nil {
		return nil, err
	}

	workdir := filepath.Join(ovlr.MountPoint, "/minifaas")

	err = os.MkdirAll(workdir, 0755)
	if err != nil {
		return nil, err
	}

	cmdexec := exec.Command("tar", "-C", workdir, "-xz")
	cmdexec.Dir = workdir
	cmdexec.Stdin = &httpBody{connR: connR}
	cmdexec.Stdout = os.Stdout
	cmdexec.Stderr = os.Stdout
	err = cmdexec.Run()
	if err != nil {
		return nil, err
	}

	err = os.Chown(workdir, config.Runner.NonRootUID, config.Runner.NonRootGID)
	if err != nil {
		return nil, err
	}

	err = os.Chown(ovlr.MountPoint, config.Runner.NonRootUID, config.Runner.NonRootGID)
	if err != nil {
		return nil, err
	}

	ip := smartdhcp.Get()
	nscmd := nsjailCommand(
		ovlr.MountPoint,
		"/minifaas",
		600,
		cmd.Env,
		"10.166."+ipStr(ip),
		"10.166.0.1",
		"tveth1",
		"255.255.0.0",
		cmd.BuildCmd...,
	)

	bb := utils.NewLimitedBuffer(512 * 1024)
	nscmd.Stdout = bb
	nscmd.Stderr = bb

	err = nscmd.Start()
	if err != nil {
		smartdhcp.Put(ip)
		ovlr.Kill()
		return nil, fmt.Errorf("%v: %v", err, bb.String())
	}

	err = nscmd.Wait()
	smartdhcp.Put(ip)
	ovlr.Umount()
	if err != nil {
		return nil, fmt.Errorf("%v: %v", err, bb.String())
	}

	codeidpath := getCodePath(cmd.CodeID)

	if utils.DirExists(codeidpath) {
		err = DelSubvolumeAbsolute(codeidpath)
		if err != nil {
			return nil, err
		}
	}

	codeidpath, err = NewSubvolume(filepath.Join("codes", cmd.CodeID), "")
	if err != nil {
		return nil, fmt.Errorf("new subvolume: %v", err)
	}

	err = utils.Copy(ovlr.UpperDir, codeidpath)
	if err != nil {
		return nil, fmt.Errorf("filesystem copy: %v", err)
	}

	ovlr.Kill()

	meta, _ := json.Marshal(map[string]interface{}{
		"date":   time.Now().Unix(),
		"image":  cmd.Image,
		"codeid": cmd.CodeID,
	})
	if utils.FileExists(filepath.Join(codeidpath, "metadata.json")) {
		err = os.Remove(filepath.Join(codeidpath, "metadata.json"))
		if err != nil {
			return nil, err
		}
	}
	err = ioutil.WriteFile(filepath.Join(codeidpath, "metadata.json"), meta, 0700)
	if err != nil {
		return nil, err
	}

	err = objectstorage.Client.PushFolderTar(codeidpath, filepath.Join("codes", cmd.CodeID))
	if err != nil {
		return nil, fmt.Errorf("objectstorage.PushFolderTar: %v", err)
	}

	logs = bb.Bytes()
	return
}

func ipStr(ip int64) string {
	end := ip & 0xFF
	return strconv.FormatInt((ip&^end)>>8, 10) + "." + strconv.FormatInt(end, 10)
}
