package utils

import (
	"io"
	"log"
	"os"

	"github.com/otiai10/copy"
)

func PathExists(path string) (bool, os.FileInfo) {
	info, err := os.Stat(path)
	if err != nil && !os.IsNotExist(err) {
		log.Fatal(err)
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return true, info
}

func FileExists(filename string) bool {
	exists, info := PathExists(filename)
	if !exists {
		return false
	}

	return !info.IsDir()
}

func DirExists(dirname string) bool {
	exists, info := PathExists(dirname)
	if !exists {
		return false
	}

	return info.IsDir()
}

func DirEmpty(dirname string) (b bool) {
	if !DirExists(dirname) {
		return
	}

	f, err := os.Open(dirname)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()

	// If the first file is EOF, the directory is empty
	if _, err = f.Readdir(1); err == io.EOF {
		b = true
	}

	return
}

// CopyFile copies both files and directories
func Copy(src string, dst string) error {
	return copy.Copy(src, dst)
}
