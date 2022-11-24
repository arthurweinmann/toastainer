package user

import (
	"net/http"
	"strconv"
	"time"

	"github.com/toastate/toastainer/internal/db/objectdb"
	"github.com/toastate/toastainer/internal/db/objectdb/objectdberror"
	"github.com/toastate/toastainer/internal/model"
	"github.com/toastate/toastainer/internal/utils"
)

type UsageResponse struct {
	Success bool                  `json:"success,omitempty"`
	Usage   *model.UserStatistics `json:"usage,omitempty"`
}

func Usage(w http.ResponseWriter, r *http.Request, userid string) {
	var month, year string

	tmp, ok := r.URL.Query()["month"]
	if ok && len(tmp) > 0 {
		month = tmp[0]
	} else {
		month = strconv.Itoa(int(time.Now().Month()))
	}

	tmp, ok = r.URL.Query()["year"]
	if ok && len(tmp) > 0 {
		year = tmp[0]
	} else {
		year = strconv.Itoa(time.Now().Year())
	}

	stats, err := objectdb.Client.GetUserStatistics(userid, month+year)
	if err != nil {
		if err == objectdberror.ErrNotFound {
			utils.SendSuccess(w, &UsageResponse{
				Success: true,
				Usage: &model.UserStatistics{
					UserID:     userid,
					Monthyear:  month + year,
					DurationMS: 0,
					CPUS:       0,
					Executions: 0,
					RAMGBS:     0,
					NetIngress: 0,
					NetEgress:  0,
				},
			})
			return
		}

		utils.SendInternalError(w, "user.Usage:objectdb.Client.GetUserStatistics", err)
		return
	}

	utils.SendSuccess(w, &UsageResponse{
		Success: true,
		Usage:   stats,
	})
}
