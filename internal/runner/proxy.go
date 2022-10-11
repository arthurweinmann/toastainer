package runner

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/toastate/toastainer/internal/config"
	"github.com/toastate/toastainer/internal/utils"
)

// ProxyCommand is to be used for established http and websocket connections
// Multipart should work by streaming the entire body without calling parsemultipartform on the LB side
type ProxyCommand struct {
	ExeID string

	ReqType    RequestType
	Headers    map[string][]string
	Trailers   map[string][]string
	Method     string
	RemoteAddr string
	IsHTTPS    bool
	URL        string // with https replaced by http and port 8080
}

type ResponseHead struct {
	Headers    map[string][]string
	Status     string
	StatusCode int
}

func proxyCommand(connR *bufio.Reader, connW *bufio.Writer) (err error) {
	defer func() {
		if err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
			err2 := writeError(connW, err)
			if err2 != nil {
				utils.Error("origin", "runner:proxyCommand", "error", fmt.Sprintf("could not write error %v: %v", err, err2))
			}
		}
	}()

	b, err := readCommand(connR)
	if err != nil {
		return err
	}

	cmd := &ProxyCommand{}
	err = json.Unmarshal(b, cmd)
	if err != nil {
		return err
	}

	err = proxyCommandInternal(cmd, connR, connW)
	if err != nil {
		return err
	}

	return nil
}

func proxyCommandInternal(cmd *ProxyCommand, connR *bufio.Reader, connW *bufio.Writer) (err error) {
	var exe *executionInProgress
	exe, err = retrieveExe(cmd.ExeID)
	if err != nil {
		return err
	}

	switch cmd.ReqType {
	case HTTPRequest:
		err = proxyHTTP(cmd, exe, connR, connW)
	case WebsocketRequest:
		err = proxyWebsocket(cmd, exe, connR, connW)
	default:
		err = fmt.Errorf("unsupported request type %v", cmd.ReqType)
		return
	}

	if err != nil {
		return err
	}

	// No need to write success here since the handles will have stream the response
	return nil
}

func retrieveExe(exeid string) (*executionInProgress, error) {
	var exe *executionInProgress
	var ok bool

	for i := 0; i < 5; i++ {
		executionsMu.RLock()
		exe, ok = executions[exeid]
		executionsMu.RUnlock()

		if ok {
			break
		}

		time.Sleep(time.Duration(250*(i+1)) * time.Millisecond)
	}

	if !ok {
		return nil, fmt.Errorf("execution %s was not found", exeid)
	}

	select {
	case <-exe.startingWait:
	case <-time.After(10 * time.Second):
		return nil, fmt.Errorf("we were not able to start execution %s in time", exeid)
	}

	if exe.errStart != nil {
		return nil, fmt.Errorf("execution %s could not be started: %v", exeid, exe.errStart)
	}

	select {
	case <-exe.exeEndNotif:
		if exe.errExe != nil {
			return nil, fmt.Errorf("Execution with id %s ended with error: %v", exe.cmd.ExeID, exe.errExe)
		} else {
			return nil, fmt.Errorf("Execution with id %s ended", exe.cmd.ExeID)
		}
	default:
	}

	return exe, nil
}

type httpBody struct {
	connR     *bufio.Reader
	remaining int
}

func (body *httpBody) Read(p []byte) (int, error) {
	n, r, err := readStreamChunk(body.connR, p, body.remaining)
	body.remaining = r

	return n, err
}

func proxyHTTP(cmd *ProxyCommand, exe *executionInProgress, connR *bufio.Reader, connW *bufio.Writer) error {
	proxyReq, err := http.NewRequest(cmd.Method, cmd.URL, &httpBody{connR: connR})
	if err != nil {
		return err
	}

	if len(cmd.Headers) > 0 {
		proxyReq.Header = http.Header(cmd.Headers)
	}
	if len(cmd.Trailers) > 0 {
		proxyReq.Trailer = http.Header(cmd.Trailers)
	}

	proxyReq.Header.Set("X-Forwarded-For", cmd.RemoteAddr)
	if cmd.IsHTTPS {
		proxyReq.Header.Set("X-Forwarded-Proto", "https")
	} else {
		proxyReq.Header.Set("X-Forwarded-Proto", "http")
	}

	resp, err := utils.ForceIPHTTPClient(exe.ip, config.Runner.ToasterPort).Do(proxyReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	rh := &ResponseHead{
		Headers:    resp.Header,
		Status:     resp.Status,
		StatusCode: resp.StatusCode,
	}

	err = writeSuccess(connW, rh)
	if err != nil {
		return err
	}

	err = StreamReader(connW, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func proxyWebsocket(cmd *ProxyCommand, exe *executionInProgress, connR *bufio.Reader, connW *bufio.Writer) error {
	h := http.Header(cmd.Headers)
	h.Set("X-Forwarded-For", cmd.RemoteAddr)
	if cmd.IsHTTPS {
		h.Set("X-Forwarded-Proto", "https")
	} else {
		h.Set("X-Forwarded-Proto", "http")
	}

	c, _, err := utils.ForceIPWebsocketDialer(exe.ip, config.Runner.ToasterPort).Dial(cmd.URL, h)
	if err != nil {
		return err
	}

	err = writeSuccess(connW, nil)
	if err != nil {
		return err
	}

	errs := make(chan error, 2)

	go func() {
		errs <- StreamReadWebsocket(c, connW)
	}()
	go func() {
		errs <- StreamWriteWebsocket(c, connR)
	}()

	err = <-errs
	c.Close()
	err = <-errs

	return err
}
