package library

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"math/rand"
	"net"
	"net/http"
	"net/http/cookiejar"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/toastate/toastcloud/internal/api/routes/toaster"
	"github.com/toastate/toastcloud/internal/api/routes/user"
	"github.com/toastate/toastcloud/internal/config"
	"github.com/toastate/toastcloud/internal/utils"
	"golang.org/x/net/publicsuffix"
)

//go:embed exampletoaster/*
var exampletoaster embed.FS

type FullAPITest struct {
	httpclientGenerator           func() *http.Client
	baseAPIURLWithoutLeadingSlash string

	hostaddr   string
	hostport   string
	hostscheme string

	toasterdomain string
	apidomain     string

	opts *FullAPITestOpts
}

type FullAPITestOpts struct {
	SetHostRedirection func(ip, hostname string) error
}

func NewFullAPITest(httpclientGenerator func() *http.Client, baseAPIURLWithoutLeadingSlash, apidomain, toasterdomain string, opts ...*FullAPITestOpts) (*FullAPITest, error) {
	fat := &FullAPITest{
		httpclientGenerator:           httpclientGenerator,
		baseAPIURLWithoutLeadingSlash: baseAPIURLWithoutLeadingSlash,
		toasterdomain:                 toasterdomain,
		apidomain:                     apidomain,
	}
	if len(opts) > 0 {
		fat.opts = opts[0]
	}

	var err error
	fat.hostscheme, fat.hostaddr, fat.hostport, err = utils.BreakBaseURL(baseAPIURLWithoutLeadingSlash)
	if err != nil {
		return nil, fmt.Errorf("invalid baseAPIURLWithoutLeadingSlash: %v", err)
	}

	if fat.hostscheme == "" {
		fat.hostscheme = "http"
	}

	return fat, nil
}

func (fat *FullAPITest) host() string {
	s := fat.hostaddr
	if fat.hostport != "" {
		s += ":" + fat.hostport
	}
	return s
}

func (fat *FullAPITest) Run() error {
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return err
	}

	dialer := &net.Dialer{
		Timeout:   60 * time.Second,
		KeepAlive: 60 * time.Second,
	}

	callclient1 := fat.httpclientGenerator()
	callclient1.Jar = jar
	callclient1.Timeout = 60 * time.Second
	if callclient1.Transport == nil {
		callclient1.Transport = &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return dialer.DialContext(ctx, network, fat.host())
			},
		}
	} else {
		switch t := callclient1.Transport.(type) {
		case *http.Transport:
			t.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
				return dialer.DialContext(ctx, network, fat.host())
			}
		default:
			return fmt.Errorf("invalid preexisting http client transport type %T, we currently support: *http.Transport", t)
		}
	}

	signinReq := &user.SigninRequest{
		Email:    "arthur.weinmann@toastate.com",
		Password: "**DONOTUSE**_ToastCloud1234@?",
	}
	signinResp := &user.SigninResponse{}
	err = fat.makeRequestWithBody(callclient1, "POST", fat.makeAPIURL("/cookiesignin"), signinReq, signinResp)
	if err == nil {
		err = fat.makeRequestWithBody(callclient1, "POST", fat.makeAPIURL("/user/deleteaccount"), &user.DeleteAccountRequest{Password: "**DONOTUSE**_ToastCloud1234@?"}, &user.DeleteAccountResponse{})
		if err != nil {
			return fmt.Errorf("/user/deleteaccount: %v", err)
		}
	} else if !strings.Contains(err.Error(), "email address not found") {
		return fmt.Errorf("invalid error for first cookiesignin call: %v", err)
	}

	signupReq := &user.SignupRequest{
		Email:    "arthur.weinmann@toastate.com",
		Password: "**DONOTUSE**_ToastCloud1234@?",
	}
	signupResp := &user.SignupResponse{}
	err = fat.makeRequestWithBody(callclient1, "POST", fat.makeAPIURL("/signup"), signupReq, signupResp)
	if err != nil {
		return fmt.Errorf("/signup: %v", err)
	}

	signinReq = &user.SigninRequest{
		Email:    "arthur.weinmann@toastate.com",
		Password: "**DONOTUSE**_ToastCloud1234@?",
	}
	signinResp = &user.SigninResponse{}
	err = fat.makeRequestWithBody(callclient1, "POST", fat.makeAPIURL("/cookiesignin"), signinReq, signinResp)
	if err != nil {
		return fmt.Errorf("/cookiesignin: %v", err)
	}

	defer func() {
		err2 := fat.makeRequestWithBody(callclient1, "POST", fat.makeAPIURL("/user/deleteaccount"), &user.DeleteAccountRequest{Password: "**DONOTUSE**_ToastCloud1234@?"}, &user.DeleteAccountResponse{})
		if err2 != nil {
			fmt.Println("could not delete account", err2)
		}
	}()

	toasterCreateReq := &toaster.CreateRequest{
		Name:                 "example1",
		Image:                "ubuntu-20.04-nodejs-golang",
		BuildCmd:             []string{"/bin/bash", "-c", `go mod init toasterexample && go get ./... && go build`},
		ExeCmd:               []string{"./toasterexample"},
		Env:                  []string{"GOPATH=/home/ubuntu/go", "GOROOT=/usr/local/go", "TERM=xterm-color", "HOME=/home/ubuntu", "PATH=/home/ubuntu/go/bin:/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"},
		JoinableForSec:       10,
		TimeoutSec:           15,
		MaxConcurrentJoiners: 10,
	}
	toasterCreateResp := &toaster.CreateResponse{}

	b, err := json.Marshal(toasterCreateReq)
	if err != nil {
		return err
	}

	direntries, _ := exampletoaster.ReadDir("exampletoaster")

	var files []fs.File
	var filepaths []string

	for i := 0; i < len(direntries); i++ {
		f, err := exampletoaster.Open(filepath.Join("exampletoaster", direntries[i].Name()))
		if err != nil {
			return err
		}
		files = append(files, f)
		filepaths = append(filepaths, direntries[i].Name())
	}

	resp, err := utils.UploadEmbedFolderMultipart(callclient1, fat.makeAPIURL("/toaster"), "POST", filepaths, files, "request", string(b))
	if err != nil {
		return err
	}
	err = fat.parseAPIResp(resp, toasterCreateResp)
	if err != nil {
		return err
	}

	if toasterCreateResp.BuildID != "" {
		for i := 0; i < 10; i++ {
			r := &toaster.GetBuildResultResponse{}
			err = fat.makeRequestWithoutBody(callclient1, "GET", fat.makeAPIURL("/toaster/build/"+toasterCreateResp.BuildID), r)
			if err != nil {
				return err
			}
			fmt.Println("################# build id check", r.InProgress, len(r.BuildLogs), len(r.BuildError), err)
			if !r.InProgress {
				if len(r.BuildError) > 0 {
					return fmt.Errorf("compilation error: %v", string(r.BuildError))
				}

				toasterCreateResp.BuildLogs = r.BuildLogs
				break
			}
			if i < 9 {
				time.Sleep(time.Duration(i+1) * 30 * time.Second)
			} else {
				return fmt.Errorf("example toaster compilation result retrieval timed out")
			}
		}
	}

	if fat.opts != nil && fat.opts.SetHostRedirection != nil {
		err = fat.opts.SetHostRedirection("127.0.0.1", toasterCreateResp.Toaster.ID+"."+fat.toasterdomain)
		if err != nil {
			return err
		}
	}

	toaster1_EXEID1, err := fat.echotoaster(callclient1, toasterCreateResp.Toaster.ID)
	if err != nil {
		return err
	}

	toaster1_EXEID2, err := fat.echotoaster(callclient1, toasterCreateResp.Toaster.ID)
	if err != nil {
		return err
	}

	if toaster1_EXEID2 != toaster1_EXEID1 {
		return fmt.Errorf("second http request joined another toaster: %v != %v", toaster1_EXEID1, toaster1_EXEID2)
	}

	sock1, err := fat.dialWebsocket(callclient1, toasterCreateResp.Toaster.ID, "/echo")
	if err != nil {
		return fmt.Errorf("could not dial test toaster through a websocket: %v", err)
	}

	for i := 0; i < 3; i++ {
		messstr := `websocket` + strconv.Itoa(i)
		err = sock1.WriteMessage(websocket.TextMessage, []byte(messstr))
		if err != nil {
			sock1.Close()
			return fmt.Errorf("could not write websocket message to toaster: %v", err)
		}

		_, mess, err := sock1.ReadMessage()
		if err != nil {
			sock1.Close()
			return fmt.Errorf("could not read websocket message from toaster: %v", err)
		}

		if string(mess) != messstr {
			sock1.Close()
			return fmt.Errorf("invalid websocket message received from toaster: %s", string(mess))
		}
	}

	err = sock1.Close()
	if err != nil {
		return fmt.Errorf("could not close websocket connection to toaster: %v", err)
	}

	time.Sleep(16 * time.Second)

	// serial join test
	var firstexeid string
	for i := 0; i < 10; i++ {
		exeid, err := fat.echotoaster(callclient1, toasterCreateResp.Toaster.ID)
		if err != nil {
			return err
		}

		if firstexeid == "" {
			firstexeid = exeid
		} else {
			if firstexeid != exeid {
				return fmt.Errorf("invalid toaster join: %v != %v", toaster1_EXEID1, toaster1_EXEID2)
			}
		}
	}

	exeid, err := fat.echotoaster(callclient1, toasterCreateResp.Toaster.ID)
	if err != nil {
		return err
	}
	if exeid == firstexeid {
		return fmt.Errorf("request could join a toaster that reached maxjoiners")
	}

	firstexeid = exeid
	for i := 0; i < 9; i++ {
		sock2, err := fat.dialWebsocket(callclient1, toasterCreateResp.Toaster.ID, "/exeid")
		if err != nil {
			return fmt.Errorf("could not dial test toaster through a websocket: %v", err)
		}

		err = sock2.WriteMessage(websocket.TextMessage, []byte("give me your exeid"))
		if err != nil {
			sock2.Close()
			return fmt.Errorf("could not write websocket message to toaster: %v", err)
		}

		_, mess, err := sock2.ReadMessage()
		if err != nil {
			sock2.Close()
			return fmt.Errorf("could not read websocket message from toaster: %v", err)
		}

		if string(mess) != firstexeid {
			sock2.Close()
			return fmt.Errorf("invalid toaster join: %v != %v", string(mess), firstexeid)
		}

		err = sock2.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(10*time.Second))
		if err != nil {
			return fmt.Errorf("could not write close to websocket connection to toaster: %v", err)
		}

		err = sock2.Close()
		if err != nil {
			return fmt.Errorf("could not close websocket connection to toaster: %v", err)
		}
	}

	time.Sleep(16 * time.Second)

	var wg sync.WaitGroup
	var mu sync.Mutex
	errs := make(chan error, 10)
	firstexeid = ""
	for i := 0; i < 10; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			exeid, err := fat.echotoaster(callclient1, toasterCreateResp.Toaster.ID)
			if err != nil {
				errs <- err
			} else {
				var ok bool
				mu.Lock()
				if firstexeid == "" {
					firstexeid = exeid
					ok = true
				} else {
					ok = exeid == firstexeid
				}
				mu.Unlock()
				if !ok {
					errs <- fmt.Errorf("one of the request did not join the same toaster: %v != %v", exeid, firstexeid)
				}
			}
		}()
	}

	wg.Wait()

	var errstr string
Fouter:
	for i := 0; i < 10; i++ {
		select {
		case err = <-errs:
			if errstr != "" {
				errstr += ", "
			}
			errstr += err.Error()
		default:
			break Fouter
		}
	}
	if errstr != "" {
		return fmt.Errorf(errstr)
	}

	return nil
}

func (fat *FullAPITest) makeAPIURL(suffix string) string {
	if suffix[0] != '/' {
		suffix = "/" + suffix
	}
	return fat.hostscheme + "://" + config.APIDomain + suffix
}

func (fat *FullAPITest) makeToasterWebsocketURL(toasterid, suffix string) string {
	if suffix[0] != '/' {
		suffix = "/" + suffix
	}

	scheme := "ws"
	if fat.hostscheme == "https" {
		scheme = "wss"
	}

	if fat.hostport != "" {
		return scheme + "://" + toasterid + "." + fat.toasterdomain + ":" + fat.hostport + suffix
	}

	return scheme + "://" + toasterid + "." + fat.toasterdomain + suffix
}

func (fat *FullAPITest) makeToasterHTTPURL(toasterid, suffix string) string {
	if suffix[0] != '/' {
		suffix = "/" + suffix
	}

	return fat.hostscheme + "://" + toasterid + "." + fat.toasterdomain + suffix
}

func (fat *FullAPITest) makeRequestWithBody(client *http.Client, method string, url string, body, resp interface{}) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(method, url, bytes.NewReader(b))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Length", strconv.Itoa(len(b)))

	fmt.Println("- Requesting ", url, "..")
	rp, err := client.Do(req)
	if err != nil {
		return err
	}

	return fat.parseAPIResp(rp, resp)
}

func (fat *FullAPITest) makeRequestWithoutBody(client *http.Client, method string, url string, resp interface{}) error {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return err
	}

	fmt.Println("- Requesting ", url, "..")
	rp, err := client.Do(req)
	if err != nil {
		return err
	}

	return fat.parseAPIResp(rp, resp)
}

func (fat *FullAPITest) parseAPIResp(rp *http.Response, resp interface{}) error {
	b, err := io.ReadAll(rp.Body)
	rp.Body.Close()

	if err != nil && err != io.EOF {
		return err
	}

	if rp.StatusCode != 200 {
		return fmt.Errorf("request status %v, error: %v", rp.StatusCode, utils.UnmarshalJSONErr(b).Message)
	}

	return json.Unmarshal(b, resp)
}

var exeindex uint32

func (fat *FullAPITest) echotoaster(client *http.Client, toasterid string) (string, error) {
	index := atomic.AddUint32(&exeindex, 1)

	rd := rand.Intn(10000)

	url := fat.makeToasterHTTPURL(toasterid, "/echo")
	fmt.Println("- Requesting ", url, "..")

	start := time.Now()
	resp, err := client.Post(url, "application/json", strings.NewReader("exampletoaster"+strconv.Itoa(rd)))
	if err != nil {
		return "", fmt.Errorf("exe %v: %v", index, err)
	}

	b, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	end := time.Now()
	fmt.Printf("    -- time to echo:%v\n", end.Sub(start))

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("exe %v: toaster error: %v", index, utils.UnmarshalJSONErr(b).Message)
	}

	if string(b) != "exampletoaster"+strconv.Itoa(rd) {
		return "", fmt.Errorf("exe %v: toaster error: returned response %s does not match echo request %s", index, string(b), "exampletoaster"+strconv.Itoa(rd))
	}

	return resp.Header.Get("X-TOASTCLOUD-EXEID"), nil
}

func (fat *FullAPITest) dialWebsocket(client *http.Client, toasterid, urlpath string) (*websocket.Conn, error) {
	d := &websocket.Dialer{
		HandshakeTimeout: 45 * time.Second,
		TLSClientConfig:  client.Transport.(*http.Transport).TLSClientConfig,
		Jar:              client.Jar,
	}
	urlpath = fat.makeToasterWebsocketURL(toasterid, urlpath)
	fmt.Println("- Requesting websocket at", urlpath, "..")
	sock, _, err := d.Dial(urlpath, nil)
	if err != nil {
		return nil, fmt.Errorf("could not dial test toaster through a websocket: %v", err)
	}

	return sock, nil
}
