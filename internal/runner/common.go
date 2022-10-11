package runner

import (
	"path/filepath"

	"github.com/toastate/toastainer/internal/config"
	"github.com/toastate/toastainer/internal/db/objectstorage"
	"github.com/toastate/toastainer/internal/utils"
)

func pullImg(image string) (string, error) {
	var err error

	pimg := filepath.Join(config.Runner.BTRFSMountPoint, "images", image)
	if !utils.DirExists(pimg) {
		pimg, err = NewSubvolume(filepath.Join("images", image), "")
		if err != nil {
			return "", err
		}
		err = objectstorage.Client.PullFolderTar(filepath.Join("images", image), pimg)
		if err != nil {
			err2 := DelSubvolumeAbsolute(pimg)
			if err2 != nil {
				utils.Error("origin", "runner:pullImg:objectstorage.Client.PullFolderTar", "error", err2)
			}
			return "", err
		}
	}

	return pimg, nil
}

func pullCode(codeid string) (string, error) {
	var err error

	pcode := getCodePath(codeid)
	if !utils.DirExists(pcode) {
		pcode, err = NewSubvolume(filepath.Join("codes", codeid), "")
		if err != nil {
			return "", err
		}
		err = objectstorage.Client.PullFolderTar(filepath.Join("codes", codeid), pcode)
		if err != nil {
			DelSubvolumeAbsolute(pcode)
			return "", err
		}
	}

	return pcode, nil
}

func getCodePath(codeid string) string {
	return filepath.Join(config.Runner.BTRFSMountPoint, "codes", codeid)
}
