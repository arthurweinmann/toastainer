package user

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/toastate/toastainer/internal/db/redisdb"
	"github.com/toastate/toastainer/internal/utils"
)

type StatsResponse struct {
	Success bool        `json:"success,omitempty"`
	Stats   *statistics `json:"statistics,omitempty"`
}

type statistics struct {
	Duration   int64   `json:"duration_ms,omitempty"`
	CPUS       int64   `json:"seconds_cpu,omitempty"`
	Executions int64   `json:"runs,omitempty"`
	RAM        float64 `json:"ram_gbs,omitempty"`
	Ingress    float64 `json:"ingress_bytes,omitempty"`
	Egress     float64 `json:"egress_bytes,omitempty"`
}

func Stats(w http.ResponseWriter, r *http.Request, userid string) {
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

	stats, err := redisdb.GetClient().HGetAll(context.Background(), "userstats_"+userid+"_"+month+year).Result()
	if err != nil {
		if err == redisdb.ErrNil {
			utils.SendSuccess(w, nil)
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

	utils.SendSuccess(w, &StatsResponse{
		Success: true,
		Stats: &statistics{
			Duration:   dms,
			CPUS:       cpus,
			Executions: executions,
			RAM:        ramgbs,
			Ingress:    ingress,
			Egress:     egress,
		},
	})
}
