package toaster

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/rs/xid"
	"github.com/toastate/toastainer/internal/api/common"
	"github.com/toastate/toastainer/internal/db/objectstorage"
	"github.com/toastate/toastainer/internal/db/redisdb"
	"github.com/toastate/toastainer/internal/model"
	"github.com/toastate/toastainer/internal/runner"
	"github.com/toastate/toastainer/internal/utils"
)

var ErrUnsuccessfulBuild = errors.New("build failed")

// buildToasterCode must be called after the toaster object is fully set since it handles saving the execution information into redis
func buildToasterCode(toaster *model.Toaster, tarpath string) (string, []byte, error) {
	var f *os.File
	f, err := os.Open(tarpath)
	if err != nil {
		return "", nil, err
	}
	defer f.Close()

	var conn net.Conn
	conn, err = runner.Connect2Any()
	if err != nil {
		return "", nil, err
	}

	connW := bufio.NewWriter(conn)
	connR := bufio.NewReader(conn)
	_, err = conn.Write([]byte{byte(runner.BuildKind)})
	if err != nil {
		return "", nil, err
	}

	cmd := &runner.BuildCommand{
		CodeID:   toaster.CodeID,
		Image:    toaster.Image,
		BuildCmd: toaster.BuildCmd,
		Env:      toaster.Env,
	}

	var b []byte
	b, err = json.Marshal(cmd)
	if err != nil {
		conn.Close()
		return "", nil, err
	}

	var payloadR []byte
	var errR error

	buildidChan := make(chan string, 1)
	wgR := make(chan bool)

	// this goroutine should not trigger before the select listens on the channel because of the networking involved in it
	// we must listen at the same time we write in case the runner needs to send us an error
	go func() {
		defer func() {
			close(wgR)

			buildid := <-buildidChan
			if buildid != "" {
				b := make([]byte, 8, 8+len(payloadR))
				binary.BigEndian.PutUint64(b, uint64(len(payloadR)))
				b = append(b, payloadR...)
				b = append(b, 0, 0, 0, 0, 0, 0, 0, 0)
				if errR != nil {
					b = append(b, 1)
					b = append(b, errR.Error()...)
				} else {
					b = append(b, 0)
				}

				err2 := objectstorage.Client.PushReader(bytes.NewReader(b), filepath.Join("buildresults", toaster.OwnerID, buildid))
				if err2 != nil {
					utils.Error("msg", "buildToasterCode:goroutibe:objectstorage.PushReader", "Error", err2)
				}

				err2 = redisdb.GetClient().Set(context.Background(), "build_"+toaster.OwnerID+buildid, "done", 1*time.Hour).Err()
				if err2 != nil {
					utils.Error("msg", "buildToasterCode:goroutibe:redis.Set", "Error", err2)
				}
				runner.PutConnection(conn)
			}
		}()

		var success bool
		success, payloadR, errR = runner.ReadResponse(connR)
		if errR != nil {
			conn.Close()
			errR = fmt.Errorf("could not read build server response: %v", errR)
			return
		}
		if !success {
			conn.Close()
			errR = ErrUnsuccessfulBuild
			return
		}

		// the exe information dump contains the toaster's codeID and not its ID
		// this allows code updates which do not require runner cache invalidations
		errR = redisdb.GetClient().Set(context.Background(), "exeinfo_"+toaster.ID, common.DumpToaterExeInfo(toaster), 0).Err()
		if err != nil {
			errR = fmt.Errorf("could not save exe info in redis: %v", errR)
			return
		}
	}()

	go func() {
		errW := runner.WriteCommand(connW, b)
		if errW == nil {
			errW = runner.StreamReader(connW, f)
			if errW != nil {
				conn.Close()
			}
		} else {
			conn.Close()
		}
	}()

	select {
	case <-wgR:
		close(buildidChan)
		runner.PutConnection(conn)
		return "", payloadR, errR

	case <-time.After(15 * time.Second):
		buildid := xid.New().String() + "_" + strconv.Itoa(int(time.Now().Unix()))

		err2 := redisdb.GetClient().Set(context.Background(), "build_"+toaster.OwnerID+buildid, "inprogress", 1*time.Hour).Err()
		if err2 != nil {
			close(buildidChan)

			return "", nil, err2
		}

		buildidChan <- buildid

		return buildid, nil, nil
	}
}
