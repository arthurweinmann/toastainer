package toaster

import (
	"context"
	"encoding/binary"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/toastate/toastcloud/internal/db/objectstorage"
	"github.com/toastate/toastcloud/internal/db/redisdb"
	"github.com/toastate/toastcloud/internal/utils"
)

type GetBuildResultRequest struct {
}

type GetBuildResultResponse struct {
	Success    bool   `json:"success,omitempty"`
	InProgress bool   `json:"in_progress,omitempty"`
	BuildLogs  []byte `json:"build_logs,omitempty"`
	BuildError []byte `json:"build_error,omitempty"`
}

// GetBuildLogs is called when a build time exceeded at least 30s
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

	fmt.Println("got build result---", string(b))

	l := binary.BigEndian.Uint64(b)
	payload := b[8 : 8+l]
	var errbuild string
	if b[8+l] > 0 {
		errbuild = string(b[9+l:])
	}

	fmt.Println("errbuild", errbuild)

	if errbuild != "" {
		utils.SendSuccess(w, &GetBuildResultResponse{
			Success:    true,
			BuildError: payload,
		})
		return
	}

	utils.SendSuccess(w, &GetBuildResultResponse{
		Success:   true,
		BuildLogs: payload,
	})
}
