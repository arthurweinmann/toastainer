package objectstorage

import (
	"fmt"
	"io"
	"net/http"

	"github.com/toastate/toastcloud/internal/config"
	localfs "github.com/toastate/toastcloud/internal/db/objectstorage/fs"
	"github.com/toastate/toastcloud/internal/db/objectstorage/s3"
)

var Client interface {
	DownloadFileInto(w http.ResponseWriter, remotePath string) error

	PushReader(f io.Reader, destPath string) error

	PushFolderTar(folder, destPath string) error
	PullFolderTar(remotePath, destination string) error

	UploadFolder(folder, dest string) error

	DeleteFolder(folder string) error
}

func Init() error {
	var err error

	switch config.ObjectStorage.Name {
	case "awss3":
		Client, err = s3.NewHandler()

	case "localfs":
		Client, err = localfs.NewHandler()

	default:
		return fmt.Errorf("not yet supported object storage: %s", config.ObjectStorage.Name)
	}

	return err
}
