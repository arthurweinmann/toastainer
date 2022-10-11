package runner

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/toastate/toastcloud/internal/utils"
)

var ErrInvalidUserID = errors.New("invalid userid")

type LogCommand struct {
	ExeID  string
	UserID string
}

func logCommand(connR *bufio.Reader, connW *bufio.Writer) (err error) {
	defer func() {
		if err != nil {
			err2 := writeError(connW, err)
			if err2 != nil {
				utils.Error("origin", "runner:logCommand", "error", fmt.Sprintf("could not write error: %v", err2))
			}
		}
	}()

	var b []byte
	b, err = readCommand(connR)
	if err != nil {
		return
	}

	cmd := &LogCommand{}
	err = json.Unmarshal(b, cmd)
	if err != nil {
		return
	}

	var logs []byte
	logs, err = logCommandInternal(cmd, connR, connW)
	if err != nil {
		return
	}

	err = writeSuccessRaw(connW, logs)
	return
}

func logCommandInternal(cmd *LogCommand, connR *bufio.Reader, connW *bufio.Writer) (logs []byte, err error) {
	var exe *executionInProgress
	exe, err = retrieveExe(cmd.ExeID)
	if err != nil {
		return
	}

	if exe.cmd.UserID != cmd.UserID {
		err = ErrInvalidUserID
		return
	}

	logs = exe.bb.CopyBytes()

	return
}
