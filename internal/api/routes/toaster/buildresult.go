package toaster

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"net/http"
	"path/filepath"

	"github.com/toastate/toastainer/internal/db/objectstorage"
	"github.com/toastate/toastainer/internal/db/redisdb"
	"github.com/toastate/toastainer/internal/model"
	"github.com/toastate/toastainer/internal/utils"
)

type GetBuildResultRequest struct {
}

type GetBuildResultResponse struct {
	Success    bool           `json:"success,omitempty"`
	InProgress bool           `json:"in_progress"`
	Toaster    *model.Toaster `json:"toaster,omitempty"`
	BuildLogs  []byte         `json:"build_logs,omitempty"`
	BuildError []byte         `json:"build_error,omitempty"`
}

// GetBuildLogs is called when a build time exceeded at least 15s
// its result will then be stored for an async retrieval because of web browser HTTP requests timeout
func GetBuildResult(w http.ResponseWriter, r *http.Request, userid, buildid string) {
	isdone, err := redisdb.GetClient().Get(context.Background(), "build_"+userid+buildid).Result()
	if err != nil {
		if err == redisdb.ErrNil {
			utils.SendError(w, "build result not found", "notFound", 404)
			return
		}

		utils.SendInternalError(w, "GetBuildResult:redis.get", err)
		return
	}

	if isdone != "done" {
		utils.SendSuccess(w, &GetBuildResultResponse{
			Success:    true,
			InProgress: true,
		})
		return
	}

	b, err := objectstorage.Client.Get(filepath.Join("buildresults", userid, buildid))
	if err != nil {
		utils.SendInternalError(w, "GetBuildResult:objectstorage.get", err)
		return
	}

	l1 := binary.BigEndian.Uint64(b[0:8])
	l2 := binary.BigEndian.Uint64(b[8:16])
	payload := b[16 : 16+l1]
	tb := b[16+l1 : 16+l1+l2]

	toaster := &model.Toaster{}
	err = json.Unmarshal(tb, toaster)
	if err != nil {
		utils.SendInternalError(w, "GetBuildResult:json.Unmarshal", err)
		return
	}

	if b[16+l1+l2] > 0 {
		utils.SendSuccess(w, &GetBuildResultResponse{
			Success:    true,
			BuildError: payload,
			Toaster:    toaster,
		})
		return
	}

	utils.SendSuccess(w, &GetBuildResultResponse{
		Success:   true,
		BuildLogs: payload,
		Toaster:   toaster,
	})
}
