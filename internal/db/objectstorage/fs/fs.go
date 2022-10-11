package fs

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/schollz/progressbar/v3"
	"github.com/toastate/toastcloud/internal/config"
	"github.com/toastate/toastcloud/internal/db/objectstorage/objectstoragerror"
	"github.com/toastate/toastcloud/internal/utils"
)

type fsHandler struct{}

func NewHandler() (*fsHandler, error) {
	h := &fsHandler{}

	// Make sure we have the permissions to write in the folder
	if !utils.DirExists(config.ObjectStorage.LocalFS.Path) {

		err := os.MkdirAll(config.ObjectStorage.LocalFS.Path, 0700)
		if err != nil {
			return nil, err
		}
	}

	testfilepath := filepath.Join(config.ObjectStorage.LocalFS.Path, "test.txt")

	f, err := os.OpenFile(testfilepath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	_, err = f.WriteString("test")
	f.Close()
	if err != nil {
		return nil, err
	}

	f, err = os.OpenFile(testfilepath, os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	b, err := io.ReadAll(f)
	f.Close()
	if err != nil {
		return nil, err
	}

	if string(b) != "test" {
		return nil, fmt.Errorf("failed write test")
	}

	err = os.Remove(testfilepath)
	if err != nil {
		return nil, err
	}

	return h, nil
}

func (h *fsHandler) Get(remotePath string) ([]byte, error) {
	b, err := os.ReadFile(filepath.Join(config.ObjectStorage.LocalFS.Path, remotePath))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, objectstoragerror.ErrNotFound
		}

		return nil, err
	}

	return b, nil
}

func (h *fsHandler) DownloadFileInto(w http.ResponseWriter, remotePath string) error {
	dest := filepath.Join(config.ObjectStorage.LocalFS.Path, remotePath)

	f, err := os.OpenFile(dest, os.O_RDWR, 0644)
	if err != nil {
		if os.IsNotExist(err) {
			return objectstoragerror.ErrNotFound
		}

		return err
	}
	defer f.Close()

	b := make([]byte, 512)
	var n, nn int
F:
	for nn < len(b) {
		n, err = f.Read(b[nn:])
		nn += n
		if err != nil {
			if err == io.EOF {
				break F
			}

			return err
		}
		if n == 0 {
			break F
		}
	}
	b = b[:nn]

	mt := http.DetectContentType(b)
	if mt == "application/octet-stream" {
		tmp := utils.MimetypeFromFileExtension(dest)
		if tmp != "" {
			mt = tmp
		}
	}

	w.Header().Set("Content-Type", mt)

	_, err = f.Seek(0, 0)
	if err != nil {
		return err
	}

	_, err = io.Copy(w, f)
	if err != nil && err != io.EOF {
		return err
	}

	return nil
}

func (h *fsHandler) PushReader(rd io.Reader, destPath string) error {
	dest := filepath.Join(config.ObjectStorage.LocalFS.Path, destPath)

	if utils.FileExists(dest) {
		err := os.Remove(dest)
		if err != nil {
			return err
		}
	} else {
		err := os.MkdirAll(filepath.Dir(dest), 0700)
		if err != nil {
			return err
		}
	}

	f, err := os.OpenFile(dest, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, rd)
	if err != nil && err != io.EOF {
		return err
	}

	return nil
}

func (h *fsHandler) PushFolderTar(folder, destPath string) error {
	dest := filepath.Join(config.ObjectStorage.LocalFS.Path, destPath)

	if utils.FileExists(dest) {
		err := os.Remove(dest)
		if err != nil {
			return err
		}
	} else {
		err := os.MkdirAll(filepath.Dir(dest), 0700)
		if err != nil {
			return err
		}
	}

	f, err := os.OpenFile(dest, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	cmd := exec.Command("tar", "cz", "./")
	cmd.Dir = folder
	bufferr := bytes.NewBuffer(nil)
	cmd.Stdout = f
	cmd.Stderr = bufferr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("%v: %v", err, bufferr.String())
	}

	return nil
}

func (h *fsHandler) PullFolderTarWithProgressBar(remotePath, destination string) error {
	return h.pullFolderTar(remotePath, destination, true)
}

func (h *fsHandler) PullFolderTar(remotePath, destination string) error {
	return h.pullFolderTar(remotePath, destination, false)
}

func (h *fsHandler) pullFolderTar(remotePath, destination string, withProgressBar bool) error {
	dest := filepath.Join(config.ObjectStorage.LocalFS.Path, remotePath)

	info, err := os.Stat(dest)
	if err != nil {
		if os.IsNotExist(err) {
			return objectstoragerror.ErrNotFound
		}

		return err
	}

	f, err := os.OpenFile(dest, os.O_RDWR, 0644)
	if err != nil {
		if os.IsNotExist(err) {
			return objectstoragerror.ErrNotFound
		}

		return err
	}
	defer f.Close()

	cmd := exec.Command("tar", "-C", destination, "-xz")

	if withProgressBar {
		progresschan := make(chan int, 1)
		defer func() {
			progresschan <- -1
		}()
		proxyr := utils.NewProxyReader(f, progresschan)
		bar := progressbar.DefaultBytes(info.Size(), "downloading "+remotePath)
		go func() {
			prevp := 0
			for {
				p := <-progresschan
				if p < 0 {
					return
				}
				bar.Add(p - prevp)
				prevp = p
			}
		}()

		cmd.Stdin = proxyr
	} else {
		cmd.Stdin = f
	}

	buf := bytes.NewBuffer(nil)
	cmd.Stdout = buf
	cmd.Stderr = buf
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("%v: %v", err, buf.String())
	}
	return nil
}

func (h *fsHandler) DeleteFolder(remotePath string) error {
	dest := filepath.Join(config.ObjectStorage.LocalFS.Path, remotePath)

	out, err := exec.Command("rm", "-rf", dest).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v: %v", err, string(out))
	}

	return nil
}

func (h *fsHandler) Exists(remotePath string) (bool, error) {
	_, err := os.Stat(remotePath)
	if err != nil && !os.IsNotExist(err) {
		return false, err
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return true, nil
}

func (h *fsHandler) UploadFolder(folder, dest string) error {
	dest = filepath.Join(config.ObjectStorage.LocalFS.Path, dest)

	if !utils.FileExists(dest) {
		err := os.MkdirAll(dest, 0700)
		if err != nil {
			return err
		}
	}

	var rels []string

	err := filepath.Walk(folder, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			rel, err := filepath.Rel(folder, path)
			if err != nil {
				return fmt.Errorf("unable to get relative path: %v %v", path, err)
			}

			file, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("failed opening file: %v %v", path, err)
			}
			defer file.Close()

			err = h.PushReader(file, filepath.Join(dest, rel))
			if err != nil {
				return err
			}

			rels = append(rels, rel)
		}

		return nil
	})

	if err != nil {
		err2 := h.DeleteFolder(dest)
		if err2 != nil {
			utils.Error("msg", "fshandler", "UploadFolder", "could not delete local file object storage following an upload folder error", err2, err)
		}

		return err
	}

	return nil
}
