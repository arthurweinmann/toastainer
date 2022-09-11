package utils

import "syscall"

func DiskUsage(path string) (all, free, used uint64, err error) {
	fs := syscall.Statfs_t{}
	err = syscall.Statfs(path, &fs)
	if err != nil {
		return
	}
	all = fs.Blocks * uint64(fs.Bsize)
	free = fs.Bfree * uint64(fs.Bsize)
	used = all - free
	return
}
