package backgroundtasks

import (
	"time"

	"github.com/toastate/toastcloud/internal/acme"
	"github.com/toastate/toastcloud/internal/utils"
)

func certificatesRoutine() {
	var lock *taskLock
	var err error

F:
	for {
		if lock != nil {
			lock.release()
			lock = nil
		}
		time.Sleep(6 * time.Hour)

		lock, err = refreshTaskLock("certificates", 1*time.Hour)
		if err != nil {
			utils.Error("msg", "certificates background routine", err)
			continue F
		}
		if lock == nil {
			continue F
		}

		builtins := acme.BuiltinCerts()
		for i := 0; i < len(builtins); i++ {
			rootDomain, err := utils.ExtractRootDomain(builtins[i][0])
			if err != nil {
				utils.Error("msg", "certificates background routine", err)
				continue F
			}

			_, _, err = acme.RetrieveCertificate(rootDomain)
			if err == acme.ErrCertificateExpired || err == acme.ErrCertificateNotFound {
				_, _, err = acme.CreateCertificate(rootDomain, builtins[i], false)
				if err != nil {
					utils.Error("msg", "certificates background routine", err)
					continue
				}
			} else if err != nil {
				utils.Error("msg", "certificates background routine", err)
				continue
			}

			if !lock.active() {
				lock, err = refreshTaskLock("certificates", 1*time.Hour)
				if err != nil {
					utils.Error("msg", "certificates background routine", err)
					continue F
				}
				if lock == nil {
					continue F
				}
			}
		}
	}
}
