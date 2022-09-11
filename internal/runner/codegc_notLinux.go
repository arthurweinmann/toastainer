//go:build !linux

package runner

import "github.com/toastate/toastcloud/internal/utils"

func runGC(gclevel int) {
	utils.Warn("msg", "codegc", "can only run on linux systems")
}
