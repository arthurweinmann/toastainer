package toaster

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/toastate/toastainer/internal/db/objectdb"
	"github.com/toastate/toastainer/internal/db/objectdb/objectdberror"
	"github.com/toastate/toastainer/internal/db/redisdb"
	"github.com/toastate/toastainer/internal/utils"
)

type RunningCountResponse struct {
	Success bool `json:"success,omitempty"`
	Count   int  `json:"count"`
}

func RunningCount(w http.ResponseWriter, r *http.Request, userid, toasterid string) {
	toaster, err := objectdb.Client.GetUserToaster(userid, toasterid)
	if err != nil {
		if err == objectdberror.ErrNotFound {
			utils.SendError(w, "could not find toaster "+toasterid, "notFound", 404)
			return
		}

		utils.SendInternalError(w, "GetCodeFile:objectdb.Client.GetUserToaster", err)
		return
	}

	count, err := redisdb.GetClient().Get(context.Background(), "toastercount_"+toaster.CodeID).Int()
	if err != nil && err != redisdb.ErrNil {
		utils.SendInternalError(w, "RunningCount:redis.Get", err)
		return
	}

	utils.SendSuccess(w, &RunningCountResponse{
		Success: true,
		Count:   count,
	})
}

type UsageResponse struct {
	Success bool   `json:"success,omitempty"`
	Usage   *usage `json:"usage"`
}

type usage struct {
	Duration   int64   `json:"duration_ms"`
	CPUS       int64   `json:"seconds_cpu"`
	Executions int64   `json:"runs"`
	RAM        float64 `json:"ram_gbs"`
	Ingress    float64 `json:"ingress_bytes"`
	Egress     float64 `json:"egress_bytes"`
}

func Usage(w http.ResponseWriter, r *http.Request, userid, toasterid string) {
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

	toaster, err := objectdb.Client.GetUserToaster(userid, toasterid)
	if err != nil {
		if err == objectdberror.ErrNotFound {
			utils.SendError(w, "could not find toaster "+toasterid, "notFound", 404)
			return
		}

		utils.SendInternalError(w, "GetCodeFile:objectdb.Client.GetUserToaster", err)
		return
	}

	stats, err := redisdb.GetClient().HGetAll(context.Background(), "toasterstats_"+toaster.CodeID+"_"+month+year).Result()
	if err != nil {
		if err == redisdb.ErrNil {
			utils.SendSuccess(w, &UsageResponse{
				Success: true,
				Usage: &usage{
					Duration:   0,
					CPUS:       0,
					Executions: 0,
					RAM:        0,
					Ingress:    0,
					Egress:     0,
				},
			})
			return
		}

		utils.SendError(w, err.Error(), "notFound", 404)
		return
	}

	dms, _ := strconv.ParseInt(stats["durationms"], 10, 64)
	cpus, _ := strconv.ParseInt(stats["cpus"], 10, 64)
	executions, _ := strconv.ParseInt(stats["executions"], 10, 64)
	ramgbs, _ := strconv.ParseFloat(stats["ramgbs"], 64)
	ingress, _ := strconv.ParseFloat(stats["ingress"], 64)
	egress, _ := strconv.ParseFloat(stats["egress"], 64)

	utils.SendSuccess(w, &UsageResponse{
		Success: true,
		Usage: &usage{
			Duration:   dms,
			CPUS:       cpus,
			Executions: executions,
			RAM:        ramgbs,
			Ingress:    ingress,
			Egress:     egress,
		},
	})
}
