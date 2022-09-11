package toaster

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/toastate/toastcloud/internal/model"
	"github.com/toastate/toastcloud/internal/runner"
)

var ErrUnsuccessfulBuild = errors.New("build failed")

func buildToasterCode(toaster *model.Toaster, tarpath string) ([]byte, error) {
	f, err := os.Open(tarpath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	conn, err := runner.Connect2Any()
	if err != nil {
		return nil, err
	}
	defer runner.PutConnection(conn)

	connW := bufio.NewWriter(conn)
	connR := bufio.NewReader(conn)
	_, err = conn.Write([]byte{byte(runner.BuildKind)})
	if err != nil {
		return nil, err
	}

	cmd := &runner.BuildCommand{
		CodeID:   toaster.CodeID,
		Image:    "ubuntu",
		BuildCmd: toaster.BuildCmd,
		Env:      toaster.Env,
	}

	b, err := json.Marshal(cmd)
	if err != nil {
		conn.Close()
		return nil, err
	}

	err = runner.WriteCommand(connW, b)
	if err != nil {
		conn.Close()
		return nil, err
	}

	err = runner.StreamReader(connW, f)
	if err != nil {
		conn.Close()
		return nil, err
	}

	success, payload, err := runner.ReadResponse(connR)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("could not read build server response: %v", err)
	}
	if !success {
		conn.Close()
		return payload, ErrUnsuccessfulBuild
	}

	return payload, nil
}
