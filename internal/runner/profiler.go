//go:build cgo
// +build cgo

package runner

/*
#include <unistd.h>
*/
import "C"

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/toastate/toastcloud/internal/utils"
)

const delayMin = 20
const delayMax = 200
const delayGrowth = 1.8

var pageSize = 0
var scClkTck = 0

func init() {
	pageSize = int(C.getpagesize())
	scClkTck = int(C.sysconf(C._SC_CLK_TCK))
}

type factProfile struct {
	RAM                 float64     // RAM in bytes * seconds
	CPU                 factCPUData // CPU in ms
	IPIn                float64     // IPIn in bytes
	IPOut               float64     // IPOut in bytes
	ProcessCreationTime int64       // ProcessCreationTime in UnixNano
	LastTick            int64       // LastTick in UnixNano
	Cmd                 *exec.Cmd
	UserID              string
	CodeID              string
	Stop                chan struct{}
}

type factSnapshot struct {
	Tick    int64
	RAM     int
	CPU     factCPUData
	IPIn    int
	IPOut   int
	Running bool
}

type factCPUData map[int]int

func (fcs factCPUData) Sum() (sum int) {
	for _, v := range fcs {
		sum += v
	}
	return
}

func (fcs factCPUData) DumpTo(fcp factCPUData) {

	for k, v := range fcp { // On traite d'abord tous les PID connus
		if k == -1 { // Cas spécial -1
			continue
		}
		if _, found := fcs[k]; found { // Si trouvé on substitue
			if fcp[k] < fcs[k] {
				fcp[k] = fcs[k]
			}
			delete(fcs, k)
		} else { // Si non trouvé, le /s process est terminé, on stocke sa valeur en tampon et on delete son pid
			fcp[-1] = v
			delete(fcp, k)
		}
	}

	for k, v := range fcs { // Tous les nouveaux PIS sont set dans fcp
		fcp[k] = v
	}
}

// Attach registers a command for profiling
func Attach(codeID, userid string, tcmd *exec.Cmd) chan struct{} {
	stop := make(chan struct{}, 1)
	fp := &factProfile{
		CPU: factCPUData{
			-1: 0,
		},
	}
	fp.UserID = userid
	fp.CodeID = codeID
	fp.Cmd = tcmd
	fp.Stop = stop

	go fp.run()

	return stop
}

func (fp *factProfile) run() {
	fp.ProcessCreationTime = time.Now().UnixNano()
	fp.LastTick = fp.ProcessCreationTime
	delayCur := time.Duration(delayMin)
	lrun := false
	for {
		closed := fp.doProfiling()
		if closed || lrun {
			break
		}

		select {
		case <-time.After(delayCur * time.Millisecond):
			if delayCur != delayMax {
				tmpD := float64(delayCur) * delayGrowth
				delayCur = time.Duration(tmpD)
				if delayCur > delayMax {
					delayCur = delayMax
				}
			}
		case <-fp.Stop: // Try a last fact tick on remaining pids
			lrun = true
		}
	}

	// Save here
	fp.writeFact()

}

func (fp *factProfile) writeFact() {
	decrQueue <- &DecrReq{
		CodeID:     fp.CodeID,
		UserID:     fp.UserID,
		Monthyear:  strconv.Itoa(utils.GetMonthYear(time.Now())),
		DurationMS: int(fp.LastTick-fp.ProcessCreationTime) / int(time.Millisecond),
		RAMGBS:     fp.RAM,
		CPUS:       fp.CPU.Sum(),
		NetIngress: fp.IPIn,
		NetEgress:  fp.IPOut,
	}
}

func (fp *factProfile) doProfiling() (closed bool) {
	defer recover() // Panic here implies execution stopping while routine is rning

	if fp.Cmd.ProcessState != nil && fp.Cmd.ProcessState.Exited() {
		return true
	}
	pid := fp.Cmd.Process.Pid
	if pid == 0 {
		return true
	}

	facturationSnapshot := makeSnapshot(pid)

	return !fp.Append(facturationSnapshot)
}

func makeSnapshot(pid int) (fs *factSnapshot) {
	var tmpA, tmpB int
	netdone := false

	fs = &factSnapshot{
		CPU: factCPUData{},
	}

	tmpA, tmpB = NetInfo(pid)

	for _, v := range Childrens(pid) {

		// Net monitoring occurs once per profiler tick
		//  and only on one of the NsJail subprocess
		if !netdone && pid != v {
			netdone = true
			tmpA, tmpB = NetInfo(v)
			if tmpA != 0 || tmpB != 0 {
				fs.IPIn = tmpA
				fs.IPOut = tmpB
			}

			if tmpA > 1<<27 || tmpB > 1<<27 {
				tmpC, tmpD := NetInfo(1)
				tmpA, tmpB = NetInfo(v) // Safe Redo
				if tmpA >= tmpC || tmpB >= tmpD {
					fs.IPIn = 0
					fs.IPOut = 0
				}
			}
		}

		// RAM = sum(master + child process ram)
		_, tmpA, _ = MemInfo(v)
		fs.RAM += tmpA

		// CPU = map[pid]cpu time
		fs.CPU[v] = CPUTime(v)

	}

	fs.Running = Running(pid)

	fs.Tick = time.Now().UnixNano()

	return fs
}

func (fp *factProfile) Append(fs *factSnapshot) (running bool) {

	fs.CPU.DumpTo(fp.CPU)

	if fs.IPIn > 0 {
		fp.IPIn = float64(fs.IPIn) / 1024 / 1024
	}

	if fs.IPOut > 0 {
		fp.IPOut = float64(fs.IPOut) / 1024 / 1024
	}

	if fs.RAM > 0 {
		intTimeDeltaNS := fs.Tick - fp.LastTick
		floatTimeDeltaSeconds := float64(intTimeDeltaNS) / (1000 * 1000 * 1000)
		ramUsage := float64(fs.RAM) * floatTimeDeltaSeconds / 1024 / 1024 / 1024
		fp.RAM += ramUsage
	}

	fp.LastTick = fs.Tick

	return fs.Running
}

// MemInfo returns info about memory for given procid
func MemInfo(procid int) (size, resident, shared int) {
	f, err := os.Open(fmt.Sprintf("/proc/%d/statm", procid))
	if err != nil {
		return
	}
	defer f.Close()
	fmt.Fscanf(f, "%d %d %d", &size, &resident, &shared)
	size *= pageSize
	resident *= pageSize
	shared *= pageSize
	return
}

// Running cheks if pid is running
func Running(pid int) bool {
	f, err := os.Open(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return false
	}
	defer f.Close()
	o := make([]byte, 64)
	f.Read(o)
	bts := bytes.Split(o, []byte{' '})
	if len(bts) < 4 {
		return false
	}
	state := bts[2][0]
	switch state {
	case 'R', 'S', 'D':
		return true
	default:
		return false
	}
}

// CPUTime returns the cpu time in MS
func CPUTime(procid int) (ms int) {
	f, err := os.Open(fmt.Sprintf("/proc/%d/stat", procid))
	if err != nil {
		return
	}
	defer f.Close()
	o := make([]byte, 512)
	f.Read(o)
	bts := bytes.Split(o, []byte{' '})
	if len(bts) < 50 {
		return
	}
	tmp, _ := strconv.Atoi(string(bts[13]))
	ms += tmp
	tmp, _ = strconv.Atoi(string(bts[14]))
	ms += tmp
	ms *= 1000
	ms /= scClkTck
	return
}

// NetInfo is used to get net i/o stats for a given process
func NetInfo(procid int) (ino, outo int) {
	f, err := os.Open(fmt.Sprintf("/proc/%d/net/netstat", procid))
	if err != nil {
		return
	}
	defer f.Close()

	o := make([]byte, 4096)
	f.Read(o)
	tmp := bytes.Split(o, []byte{'\n'})[3]
	if len(tmp) < 4 {
		return
	}
	bts := bytes.Split(tmp, []byte{' '})
	if len(bts) < 9 {
		return
	}
	ino, _ = strconv.Atoi(string(bts[7]))
	outo, _ = strconv.Atoi(string(bts[8]))
	return
}

// Childrens returns all children pids recursively
func Childrens(process int) []int {
	o := map[int]interface{}{
		process: nil,
	}
	childsR(process, o)
	out := make([]int, len(o))
	i := 0
	for k := range o {
		out[i] = k
		i++
	}
	return out
}

func childsR(process int, o map[int]interface{}) {
	base := "/proc/" + strconv.Itoa(process) + "/task"
	finf, _ := ioutil.ReadDir(base)
	for _, v := range finf {
		a, _ := ioutil.ReadFile(base + "/" + v.Name() + "/children")
		if len(a) == 0 {
			continue
		}
		psl := bytes.Split(a, []byte{' '})
		for _, v := range psl {
			a, e := strconv.Atoi(string(v))
			if e != nil {
				continue
			}
			if a != 0 {
				if _, fnd := o[a]; fnd {
					continue
				}
				o[a] = nil
				childsR(a, o)
			}
		}
	}
}
