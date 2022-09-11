package atomicpath

import (
	"runtime"

	"github.com/toastate/toastcloud/internal/utils"
)

func GetLockShards() int {
	lockShard := runtime.NumCPU()
	if lockShard == 0 {
		panic("runtime.NumCPU() == 0")
	}
	if lockShard == 1 {
		lockShard = 2
	}
	if !utils.IsPowerOf2(lockShard) {
		lockShard = int(utils.NextHighestPowerOf2(uint32(lockShard)))
	}

	return lockShard
}
