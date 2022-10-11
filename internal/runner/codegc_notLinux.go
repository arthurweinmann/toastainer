//go:build !linux

package runner

import "github.com/toastate/toastainer/internal/utils"

func runGC(gclevel int) {
	utils.Warn("msg", "codegc", "can only run on linux systems")
}
