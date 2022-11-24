package runner

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/toastate/toastainer/internal/config"
	"github.com/toastate/toastainer/internal/runner/smartdhcp"
	"github.com/toastate/toastainer/internal/utils"
)

var executions = map[string]*executionInProgress{}
var executionsMu sync.RWMutex

var executionWG sync.WaitGroup

type executionInProgress struct {
	cmd *ExecutionCommand

	startingWait chan bool
	errStart     error // after waiting on startingWait, you can check err

	exeEndNotif chan bool
	errExe      error

	pexe string

	ip                string
	ipLastPart        int64
	ipReleaseDemanded bool
	ipReleaseMu       sync.Mutex

	ovlr  *OverlayDir
	nscmd *exec.Cmd
	bb    *utils.LimitedBuffer

	profilerClose chan struct{}
}

type ExecutionCommand struct {
	ExeID      string
	CodeID     string
	UserID     string
	Image      string
	ExeCmd     []string
	Env        []string
	TimeoutSec int
}

func executeCommand(connR *bufio.Reader, connW *bufio.Writer) (err error) {
	defer func() {
		if err != nil {
			err2 := writeError(connW, err)
			if err2 != nil {
				utils.Error("origin", "runner:executeCommand", "error", fmt.Sprintf("could not write error: %v", err2))
			}
		}
	}()

	var b []byte
	b, err = readCommand(connR)
	if err != nil {
		return
	}

	cmd := &ExecutionCommand{}
	err = json.Unmarshal(b, cmd)
	if err != nil {
		return
	}

	_, err = executeCommandInternal(cmd)
	if err != nil {
		return
	}

	err = writeSuccess(connW, nil)
	return
}

func executeCommandInternal(cmd *ExecutionCommand) (exe *executionInProgress, err error) {
	cmd.Env = append(cmd.Env, "TOASTATE_EXEID="+cmd.ExeID, "TOASTATE_CODEID="+cmd.CodeID)

	exe = &executionInProgress{
		cmd:          cmd,
		startingWait: make(chan bool),
		exeEndNotif:  make(chan bool),
	}

	executionsMu.Lock()

	// TODO: remove when the system is stable enough that we are sure not to need this check
	tmp, ok := executions[cmd.ExeID]
	if ok {
		executionsMu.Unlock()
		exe = tmp
		return
	}

	executions[cmd.ExeID] = exe

	executionsMu.Unlock()

	err = exe.start()
	if err != nil {
		return
	}

	return
}

func (exe *executionInProgress) start() (err error) {
	defer func() {
		exe.errStart = err
		close(exe.startingWait)
	}()

	var pimg string
	pimg, err = pullImg(exe.cmd.Image)
	if err != nil {
		return fmt.Errorf("pullImg %s: %v", exe.cmd.Image, err)
	}

	var pcode string
	pcode, err = pullCode(exe.cmd.CodeID)
	if err != nil {
		return fmt.Errorf("pullCode: %v", err)
	}

	updateChtimes <- filepath.Join(pcode, "metadata.json")

	select {
	case exe.pexe = <-btrfsPool:
	case <-time.After(10 * time.Second):
		return fmt.Errorf("could not create execution volume")
	}

	exe.ovlr = NewOverlayDir(filepath.Base(exe.pexe))
	err = exe.ovlr.Mount([]string{pcode, pimg})
	if err != nil {
		return fmt.Errorf("overlay mount: %v", err)
	}

	err = os.Chown(exe.ovlr.MountPoint, config.Runner.NonRootUID, config.Runner.NonRootGID)
	if err != nil {
		return fmt.Errorf("chown: %v", err)
	}

	exe.ipLastPart = smartdhcp.Get()
	exe.ip = "10.166." + ipStr(exe.ipLastPart)

	exe.nscmd = nsjailCommand(
		exe.ovlr.MountPoint,
		"/toastainer",
		exe.cmd.TimeoutSec,
		exe.cmd.Env,
		exe.ip,
		"10.166.0.1",
		"tveth1",
		"255.255.0.0",
		maxMemoryPerToasterMega,
		exe.cmd.ExeCmd...,
	)

	exe.bb = utils.NewLimitedBuffer(512 * 1024)
	if config.LogLevel == "debug" || config.LogLevel == "all" {
		t := utils.NewTeeWriter(exe.bb, nsjaillogs)
		exe.nscmd.Stdout = t
		exe.nscmd.Stderr = t
	} else {
		exe.nscmd.Stdout = exe.bb
		exe.nscmd.Stderr = exe.bb
	}

	executionWG.Add(1)
	err = exe.nscmd.Start()
	if err != nil {
		executionWG.Done()
		err = fmt.Errorf("could not start execution: %v: %v", err, exe.bb.String())
		return
	}
	exe.profilerClose = Attach(exe.cmd.CodeID, exe.cmd.UserID, exe.nscmd)
	go exe.wait() // executionWG.Done() when this goroutine returns

	select {
	case <-exe.exeEndNotif:
		err = fmt.Errorf("execution prematurely ended: %v", exe.errExe)
		return
	case <-time.After(50 * time.Millisecond):
	}

	var conn net.Conn
	var iterwait int
	for {
		conn, err = net.DialTimeout("tcp", exe.ip+":8080", 5*time.Second)
		if err != nil {
			iterwait++
			if iterwait > 10 {
				exe.nscmd.Process.Kill()
				err = fmt.Errorf("could not start execution %s: %v", exe.cmd.ExeID, exe.bb.String())
				return
			}

			select {
			case <-exe.exeEndNotif:
				err = fmt.Errorf("execution prematurely ended: %v", exe.errExe)
				return
			case <-time.After(500 * time.Millisecond):
			}

			continue
		}
		if conn != nil {
			conn.Close()
			break
		}
	}

	return nil
}

func (exe *executionInProgress) wait() {
	var err error

	incrQueue <- exe.cmd.CodeID

	defer func() {
		exe.errExe = err
		close(exe.exeEndNotif)

		// profiler write in decrQueue
		close(exe.profilerClose)

		executionWG.Done()

		// GC delay
		// Make sure requests intended for one execution are not sent to another execution that got its IP
		// Serve potential exe error or that it ended to the last requests
		time.Sleep(time.Duration(exe.cmd.TimeoutSec) * time.Second)

		smartdhcp.Put(exe.ipLastPart)

		executionsMu.Lock()
		delete(executions, exe.cmd.ExeID)
		executionsMu.Unlock()
	}()

	err = exe.nscmd.Wait()
	if err != nil {
		err = fmt.Errorf("%v: %v", err, exe.bb.String())
	}

	// Cleanup
	// exe ip is put back into pool when LB calls GC command on it
	exe.ovlr.Kill()
	select {
	case btrfsDelPool <- exe.pexe:
	default:
		tmperr := DelSubvolumeAbsolute(exe.pexe)
		if tmperr != nil {
			fmt.Println("could not delete subvolume", tmperr)
		}
	}

	return
}
