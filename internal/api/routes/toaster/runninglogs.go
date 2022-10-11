package toaster

import (
	"bufio"
	"context"
	"encoding/json"
	"net"
	"net/http"

	"github.com/toastate/toastainer/internal/db/redisdb"
	"github.com/toastate/toastainer/internal/runner"
	"github.com/toastate/toastainer/internal/utils"
)

type GetRunningLogsResponse struct {
	Success bool   `json:"success"`
	Logs    []byte `json:"logs"`
}

// toasterid is contained within the exeid
func GetRunningLogs(w http.ResponseWriter, r *http.Request, userid, exeid string) {
	tvsipstr, err := redisdb.GetClient().Get(context.Background(), exeid).Result()
	if err != nil {
		if err == redisdb.ErrNil {
			utils.SendError(w, "could not find toaster", "notFound", 404)
			return
		}
	}

	runnerip := net.ParseIP(tvsipstr).To4()

	conn, err := runner.Connect2(runnerip)
	if err != nil {
		utils.SendInternalError(w, "GetRunningLogs:runner.Connect2", err)
		return
	}
	defer runner.PutConnection(conn)

	connW := bufio.NewWriter(conn)
	connR := bufio.NewReader(conn)

	cmd := &runner.LogCommand{
		UserID: userid, // the runner checks that the execution is from a toaster owned by this user and throws an error if it is not
		ExeID:  exeid,
	}

	b, err := json.Marshal(cmd)
	if err != nil {
		conn.Close()
		utils.SendInternalError(w, "GetRunningLogs:json.Marshal", err)
		return
	}

	err = runner.WriteCommand(connW, b)
	if err != nil {
		conn.Close()
		utils.SendInternalError(w, "GetRunningLogs:runner.WriteCommand", err)
		return
	}

	var success bool
	var payload []byte
	success, payload, err = runner.ReadResponse(connR)
	if err != nil {
		conn.Close()
		utils.SendInternalError(w, "GetRunningLogs:runner.ReadResponse", err)
		return
	}
	if !success {
		conn.Close()
		if string(payload) == runner.ErrInvalidUserID.Error() {
			utils.SendError(w, "could not find toaster", "notFound", 404)
			return
		}

		utils.SendInternalError(w, "GetRunningLogs:runner.ReadResponse", err)
		return
	}

	utils.SendSuccess(w, &GetRunningLogsResponse{
		Success: true,
		Logs:    payload,
	})
}
