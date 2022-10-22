package objectstorage

import (
	"fmt"
	"io"
	"net/http"

	"github.com/toastate/toastainer/internal/config"
	localfs "github.com/toastate/toastainer/internal/db/objectstorage/fs"
	"github.com/toastate/toastainer/internal/db/objectstorage/s3"
)

var Client interface {
	DownloadFileInto(w http.ResponseWriter, remotePath string) error

	PushReader(f io.Reader, destPath string) error

	Get(remotePath string) ([]byte, error)

	PushFolderTar(folder, destPath string) error

	PullFolderTar(remotePath, destination string) error
	PullFolderTarWithProgressBar(remotePath, destination string) error

	UploadFolder(folder, dest string) error

	Exists(remotePath string) (bool, error)

	DeleteFolder(folder string) error
}

func Init() error {
	var err error

	switch config.ObjectStorage.Name {
	case "awss3":
		Client, err = s3.NewHandler()

	case "localfs":
		if config.NodeDiscovery {
			return fmt.Errorf("you cannot use local filesystem for objectstorage when in a multi-node architecture")
		}
		Client, err = localfs.NewHandler()

	default:
		return fmt.Errorf("not yet supported object storage: %s", config.ObjectStorage.Name)
	}

	return err
}
