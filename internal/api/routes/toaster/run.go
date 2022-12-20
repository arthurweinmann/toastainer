package toaster

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/toastate/toastainer/internal/api/common"
	"github.com/toastate/toastainer/internal/config"
	"github.com/toastate/toastainer/internal/db/redisdb"
	"github.com/toastate/toastainer/internal/model"
	"github.com/toastate/toastainer/internal/nodes"
	"github.com/toastate/toastainer/internal/runner"
	"github.com/toastate/toastainer/internal/utils"
)

var upgrader = websocket.Upgrader{
	HandshakeTimeout: 5 * time.Second,
	Error: func(w http.ResponseWriter, r *http.Request, status int, reason error) {
		utils.SendError(w, reason.Error(), "websocketProtocol", status)
	},
}

func RunToaster(w http.ResponseWriter, r *http.Request, exeid, toasteridOrSubdomain, requestedRegion string, isHTTPS bool) {
	var tvsIP net.IP
	var toaster *model.Toaster
	var err error

	if exeid == "" {
		b, err := redisdb.GetClient().Get(context.Background(), "exeinfo_"+toasteridOrSubdomain).Bytes()
		if err != nil {
			if err == redisdb.ErrNil {
				utils.SendError(w, "could not find toaster", "notFound", 404)
				return
			}

			utils.SendInternalError(w, "RunToaster:redisdb.Get", err)
			return
		}

		toaster = common.ParseToasterExeInfo(b)

		exeid, tvsIP, err = toggleNewExecution(toaster)
		if err != nil {
			utils.SendInternalError(w, "RunToaster:toggleNewExecution", fmt.Errorf("codeid: %v, err: %v", toaster.CodeID, err))
			return
		}
	} else {
		tvsipstr, err := redisdb.GetClient().Get(context.Background(), exeid).Result()

		if err != nil && err != redisdb.ErrNil {
			utils.SendInternalError(w, "RunToaster:redisdb.Get", err)
			return
		}

		if err == nil {
			tvsIP = net.ParseIP(tvsipstr).To4()
		} else {
			tvsIP, isnew, err := forceJoinExecution(exeid, 300)
			if err != nil {
				utils.SendInternalError(w, "RunToaster:forceJoinExecution", err)
				return
			}
			if isnew {
				b, err := redisdb.GetClient().Get(context.Background(), "exeinfo_"+toasteridOrSubdomain).Bytes()
				if err != nil {
					if err == redisdb.ErrNil {
						utils.SendError(w, "could not find toaster", "notFound", 404)
						return
					}

					utils.SendInternalError(w, "RunToaster:redisdb.Get", err)
					return
				}

				toaster = common.ParseToasterExeInfo(b)

				err = startExecution(exeid, toaster.OwnerID, tvsIP, toaster)
				if err != nil {
					utils.SendInternalError(w, "RunToaster:startExecution", err)
					return
				}
			}
		}
	}

	err = proxyToasterRequest(w, r, exeid, toasteridOrSubdomain, requestedRegion, isHTTPS, tvsIP)
	if err != nil {
		utils.SendInternalError(w, "RunToaster:proxyToasterRequest", err)
		return
	}
}

func startExecution(exeid, userid string, runnerip net.IP, toaster *model.Toaster) error {
	var err error
	var conn net.Conn

	defer func() {
		if err != nil {
			terr := redisdb.DelayedForceRefreshAutojoin(toaster.CodeID, 30*time.Second)
			if terr != nil {
				fmt.Println("ERROR startExecution redis.DelayedForceRefreshAutojoin:", terr)
			}
		}
	}()

	conn, err = runner.Connect2(runnerip)
	if err != nil {
		return err
	}
	defer runner.PutConnection(conn)

	connW := bufio.NewWriter(conn)
	connR := bufio.NewReader(conn)
	_, err = conn.Write([]byte{byte(runner.ExecuteKind)})
	if err != nil {
		conn.Close()
		return err
	}

	cmd := &runner.ExecutionCommand{
		CodeID:     toaster.CodeID,
		Image:      toaster.Image,
		ExeID:      exeid,
		UserID:     userid,
		ExeCmd:     toaster.ExeCmd,
		TimeoutSec: toaster.TimeoutSec,
		Env:        append(toaster.Env, "TOASTAINER_EXE_ID="+exeid),
	}
	b, err := json.Marshal(cmd)
	if err != nil {
		conn.Close()
		return err
	}

	err = runner.WriteCommand(connW, b)
	if err != nil {
		conn.Close()
		return err
	}

	success, payload, err := runner.ReadResponse(connR)
	if err != nil {
		conn.Close()
		return fmt.Errorf("could not read execution server response: %v", err)
	}
	if !success {
		conn.Close()
		return fmt.Errorf("unsuccessful execution: %v", string(payload))
	}

	return nil
}

func generateExeID() (string, error) {
	exeid, err := utils.UniqueSecureID60()
	if err != nil {
		return "", err
	}
	return "ex_" + exeid, nil
}

func toggleNewExecution(toaster *model.Toaster) (string, net.IP, error) {
	tmpexeid, err := generateExeID()
	if err != nil {
		return "", nil, err
	}
	tmpexeid = "ex_" + tmpexeid

	tmptvsip := nodes.PickTVS()
	tmptvsipstr := tmptvsip.String()

	exeid, tvsipstr, err := redisdb.JoinOrCreateExecution(toaster.CodeID, tmpexeid, tmptvsipstr, toaster.MaxConcurrentJoiners, toaster.JoinableForSec, toaster.TimeoutSec)
	if err != nil {
		return "", nil, err
	}

	if exeid != tmpexeid {
		return exeid, net.ParseIP(tvsipstr).To4(), nil
	} else if tvsipstr != tmptvsipstr {
		panic("should not happen")
	}

	err = startExecution(exeid, toaster.OwnerID, tmptvsip, toaster)
	if err != nil {
		return "", nil, err
	}

	return tmpexeid, tmptvsip, nil
}

func forceJoinExecution(exeid string, forceExetimeoutSec int) (net.IP, bool, error) {
	tmptvsip := nodes.PickTVS()
	tmptvsipstr := tmptvsip.String()

	tvsipstr, isnew, err := redisdb.GetOrForceExeIDExecution(exeid, tmptvsipstr, forceExetimeoutSec)
	if err != nil {
		return nil, false, err
	}

	if tvsipstr == tmptvsipstr {
		return tmptvsip, isnew, nil
	}

	return net.ParseIP(tvsipstr).To4(), isnew, nil
}

func proxyToasterRequest(w http.ResponseWriter, r *http.Request, exeid, toasteridOrSubdomain, requestedRegion string, isHTTPS bool, runnerip net.IP) error {
	iswebsocket := utils.IsWebSocketRequest(r)
	var websock *websocket.Conn

	conn, err := runner.Connect2(runnerip)
	if err != nil {
		return err
	}
	defer runner.PutConnection(conn)

	connW := bufio.NewWriter(conn)
	connR := bufio.NewReader(conn)

	_, err = conn.Write([]byte{byte(runner.ProxyKind)})
	if err != nil {
		conn.Close()
		return err
	}

	var p string
	if iswebsocket {
		p = "ws://" + utils.StripPort(r.URL.Host) + ":8080" + r.URL.Path
	} else {
		p = "http://" + utils.StripPort(r.URL.Host) + ":8080" + r.URL.Path
		if r.URL.RawQuery != "" {
			p += "?" + r.URL.RawQuery
		}
		if r.URL.RawFragment != "" {
			p += "#" + r.URL.RawFragment
		}
	}

	cmd := &runner.ProxyCommand{
		ExeID:      exeid,
		Headers:    r.Header,
		Trailers:   r.Trailer,
		Method:     r.Method,
		RemoteAddr: r.RemoteAddr,
		IsHTTPS:    isHTTPS,
		URL:        p,
	}

	if iswebsocket {
		websock, err = upgrader.Upgrade(w, r, nil)
		if err != nil {
			conn.Close()
			return err
		}

		cmd.ReqType = runner.WebsocketRequest

		delete(cmd.Headers, "Upgrade")
		delete(cmd.Headers, "Connection")
		delete(cmd.Headers, "Sec-Websocket-Key")
		delete(cmd.Headers, "Sec-Websocket-Version")
		delete(cmd.Headers, "Sec-Websocket-Extensions")
	} else {
		cmd.ReqType = runner.HTTPRequest
	}

	var b []byte
	b, err = json.Marshal(cmd)
	if err != nil {
		if websock != nil {
			websock.Close()
		}
		conn.Close()
		return err
	}

	err = runner.WriteCommand(connW, b)
	if err != nil {
		if websock != nil {
			websock.Close()
		}
		conn.Close()
		return err
	}

	var success bool
	var payload []byte

	if iswebsocket {
		success, payload, err = runner.ReadResponse(connR)
		if err != nil {
			if websock != nil {
				websock.Close()
			}
			conn.Close()
			return fmt.Errorf("could not read execution server response: %v", err)
		}
		if !success {
			if websock != nil {
				websock.Close()
			}
			conn.Close()
			err = fmt.Errorf("unsuccessful execution: %v", string(payload))
			return err
		}

		errs := make(chan error, 2)

		go func() {
			errs <- runner.StreamReadWebsocket(websock, connW)
		}()
		go func() {
			errs <- runner.StreamWriteWebsocket(websock, connR)
		}()

		err = <-errs
		websock.Close()

		if err != nil && err != io.EOF && websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseNoStatusReceived, websocket.CloseAbnormalClosure) {
			conn.Close()
			return err
		}
	} else {
		err = runner.StreamReader(connW, r.Body)
		if err != nil {
			conn.Close()
			return err
		}

		success, payload, err = runner.ReadResponse(connR)
		if err != nil {
			conn.Close()
			return fmt.Errorf("could not read execution server response: %v", err)
		}
		if !success {
			conn.Close()
			err = fmt.Errorf("unsuccessful execution: %v", string(payload))
			return err
		}

		rh := &runner.ResponseHead{}
		err = json.Unmarshal(payload, rh)
		if err != nil {
			conn.Close()
			return fmt.Errorf("could not read execution server response: %v", err)
		}

		// SECURITY: we need to intercept set-cookie headers from response and make sure the domain is the subdomain of the toaster so cookie cannot be stolen by others
		cks := utils.ReadSetCookies(http.Header(rh.Headers))
		for i := 0; i < len(cks); i++ {
			for j := 0; j < len(cks[i]); j++ {
				if requestedRegion == "" {
					cks[i][j].Domain = toasteridOrSubdomain + "." + config.ToasterDomain
				} else {
					cks[i][j].Domain = toasteridOrSubdomain + "." + requestedRegion + "." + config.ToasterDomain
				}
			}
		}
		if len(cks) > 0 {
			rh.Headers["Set-Cookie"] = utils.DumpCookie(cks)
		}

		for n, vs := range rh.Headers {
			for i := 0; i < len(vs); i++ {
				w.Header().Add(n, vs[i])
			}
		}
		w.Header().Set("X-TOASTAINER-EXEID", exeid)

		w.WriteHeader(rh.StatusCode)

		err = runner.PipeReadStream(connR, w)
		if err != nil {
			conn.Close()
			return fmt.Errorf("could not read execution server response: %v", err)
		}
	}

	return nil
}
