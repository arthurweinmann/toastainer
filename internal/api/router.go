package api

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/toastate/toastainer/internal/acme"
	"github.com/toastate/toastainer/internal/api/auth"
	"github.com/toastate/toastainer/internal/api/dynamicroutes"
	"github.com/toastate/toastainer/internal/api/routes/subdomain"
	"github.com/toastate/toastainer/internal/api/routes/toaster"
	userroute "github.com/toastate/toastainer/internal/api/routes/user"
	"github.com/toastate/toastainer/internal/config"
	"github.com/toastate/toastainer/internal/utils"
)

type Router struct {
	isHTTPS bool
}

func NewHTTPRouter() *Router {
	return &Router{}
}

func NewHTTPSRouter() *Router {
	return &Router{
		isHTTPS: true,
	}
}

func (s *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	stripedhost := utils.StripPort(r.Host)

	if strings.HasPrefix(r.URL.Path, acme.ACME_CHALLENGE_URL_PREFIX) && len(r.URL.Path) > len(acme.ACME_CHALLENGE_URL_PREFIX) {
		keyauth, err := acme.GetChallenge(stripedhost, r.URL.Path[len(acme.ACME_CHALLENGE_URL_PREFIX):])
		if err != nil {
			fmt.Println("certificates.GetChallenge", err)
			w.WriteHeader(404)
			return
		}

		w.WriteHeader(200)
		w.Write(keyauth)

		return
	}

	if stripedhost == config.APIDomain {
		if !s.isHTTPS {
			utils.Redirect2HTTPS(w, r)
			return
		}

		s.api(w, r)
		return
	}

	if stripedhost == config.DashboardDomain {
		if !s.isHTTPS {
			utils.Redirect2HTTPS(w, r)
			return
		}

		s.dashboard(w, r)
		return
	}

	s.proxy2Toaster(w, r)
}

func (s *Router) dashboard(w http.ResponseWriter, r *http.Request) {
	// Check for .. in the path and respond with an error if it is present
	// otherwise users could access any file on the server
	if utils.ContainsDotDot(r.URL.Path) {
		// TODO: send web dedicated error page
		utils.SendError(w, "invalid Path", "invalidPath", 400)
		return
	}

	setupCORS(w, r.Header.Get("Origin"))

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	upath := r.URL.Path
	if !strings.HasPrefix(upath, "/") {
		upath = "/" + upath
		r.URL.Path = upath
	}

	const indexPage = "index.html"

	fullName := filepath.Join(config.Home, "web", filepath.FromSlash(path.Clean(upath)))

	if fullName[len(fullName)-1] == '/' {
		fullName = filepath.Join(fullName, indexPage)
	}

	info, err := os.Stat(fullName)

	valid := false
	if err != nil || info.IsDir() {
		if err != nil && !os.IsNotExist(err) {
			utils.SendInternalError(w, "router:dashboard", err)
			return
		}

		info, err = os.Stat(fullName + ".html")
		if err != nil || info.IsDir() {
			if err != nil && !os.IsNotExist(err) {
				utils.SendInternalError(w, "router:dashboard", err)
				return
			}

			info, err := os.Stat(filepath.Join(fullName, indexPage))
			if err != nil || info.IsDir() {
				if err != nil && !os.IsNotExist(err) {
					utils.SendInternalError(w, "router:dashboard", err)
					return
				}
			} else {
				fullName = filepath.Join(fullName, indexPage)
				valid = true
			}
		} else {
			fullName = fullName + ".html"
			valid = true
		}
	} else {
		valid = true
	}

	if !valid {
		// TODO: use web 404 dedicated page
		utils.SendError(w, "page not found", "notFound", 404)
		return
	}

	content, err := os.Open(fullName)
	if err != nil {
		utils.SendInternalError(w, "router:dashboard", err)
		return
	}

	ctype := mime.TypeByExtension(filepath.Ext(fullName))
	if ctype == "" {
		var buf [512]byte
		n, _ := io.ReadFull(content, buf[:])
		ctype = http.DetectContentType(buf[:n])

		var nn int
		for nn < n {
			l, err := w.Write(buf[nn:])
			nn += l
			if err != nil {
				utils.SendInternalError(w, "router:dashboard", err)
				return
			}
		}
	}

	w.Header().Set("Content-Type", ctype)
	io.Copy(w, content)
}

type rootResponse struct {
	Success bool `json:"success"`
}

func (s *Router) api(w http.ResponseWriter, r *http.Request) {
	spath := utils.SplitSlash(r.URL.Path)

	setupCORS(w, r.Header.Get("Origin"))

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if len(spath) == 0 {
		utils.SendSuccess(w, &rootResponse{
			Success: true,
		})
		return
	}

	switch r.Header.Get("X-TOASTAINER-APIVERSION") {
	case "", "v1":
		switch spath[0] {
		case "signup":
			userroute.Signup(w, r)
			return
		case "signin":
			userroute.Signin(w, r)
			return
		case "cookiesignin":
			userroute.CookieSignin(w, r)
			return
		case "forgotten-password":
			userroute.ForgottenPasswordSendLink(w, r)
			return
		case "reset-password":
			userroute.ForgottenPasswordReset(w, r)
			return
		default:
			dynRoute := dynamicroutes.GetDynamicRoute(spath)
			if dynRoute != nil {
				dynRoute(w, r)
				return
			}
		}
	}

	// Authentication checkpoint
	user, sessToken, continu := auth.Auth(w, r)
	if !continu {
		return
	}

	switch r.Header.Get("X-TOASTAINER-APIVERSION") {
	case "", "v1":
		switch spath[0] {
		case "user":
			if len(spath) > 1 {
				switch spath[1] {
				case "usage":
					userroute.Usage(w, r, user.ID)
					return
				case "changepassword":
					userroute.UpdatePassword(w, r, user.ID)
					return
				case "picture":
					switch r.Method {
					case "POST":
						userroute.UpdateUserPictureRoute(w, r, user.ID)
						return
					case "GET":
						if len(spath) > 3 {
							userroute.GetUserPicture(w, r, user.ID, spath[2], spath[3])
							return
						}
					}

				case "signout":
					userroute.Signout(w, r, user.ID, sessToken)
					return
				case "deleteaccount":
					userroute.DeleteAccount(w, r, user.ID, sessToken)
					return
				}
			} else {
				switch r.Method {
				case "GET":
					userroute.GetUser(w, r, user.ID)
					return
				}
			}
		case "toaster":
			switch r.Method {
			case "POST":
				if len(spath) > 1 {
					switch spath[1] {
					case "picture":
						if len(spath) > 2 {
							toaster.UpdateToasterPictureRoute(w, r, user.ID, spath[2])
							return
						}
					}
				} else {
					toaster.Create(w, r, user.ID)
					return
				}
			case "PUT":
				if len(spath) > 1 {
					toaster.Update(w, r, user.ID, spath[1])
					return
				}
			case "GET":
				if len(spath) == 2 {
					switch spath[1] {
					case "list":
						toaster.List(w, r, user.ID)
						return
					default:
						toaster.Get(w, r, user.ID, spath[1])
						return
					}
				} else if len(spath) > 2 {
					switch spath[1] {
					case "picture":
						if len(spath) > 3 {
							toaster.GetToasterPicture(w, r, user.ID, spath[2], spath[3])
							return
						}
					case "build":
						toaster.GetBuildResult(w, r, user.ID, spath[2])
						return
					case "livelogs":
						toaster.GetRunningLogs(w, r, user.ID, spath[2])
						return
					case "file":
						toaster.GetCodeFile(w, r, user.ID, spath[2], strings.Join(spath[3:], "/"))
						return
					case "runningcount":
						toaster.RunningCount(w, r, user.ID, spath[2])
						return
					case "usage":
						toaster.Usage(w, r, user.ID, spath[2])
						return
					}
				}
			case "DELETE":
				if len(spath) > 1 {
					toaster.Delete(w, r, user.ID, spath[1])
					return
				}
			}
		case "subdomain":
			switch r.Method {
			case "POST":
				subdomain.Create(w, r, user.ID)
				return
			case "PUT":
				if len(spath) > 1 {
					subdomain.Update(w, r, user.ID, spath[1])
					return
				}
			case "GET":
				if len(spath) > 1 {
					switch spath[1] {
					case "list":
						subdomain.List(w, r, user.ID)
						return
					default:
						subdomain.Get(w, r, user.ID, spath[1])
						return
					}
				}
			case "DELETE":
				if len(spath) > 1 {
					subdomain.Delete(w, r, user.ID, spath[1])
					return
				}
			}
		}
	}

	utils.SendError(w, fmt.Sprintf("Route %s with method %s does not exist", r.URL.Path, r.Method), "invalidRoute", 404)
}

func (s *Router) proxy2Toaster(w http.ResponseWriter, r *http.Request) {
	domain := utils.StripPort(r.Host)

	var exeid string

	// Custom Domain
	if !strings.HasSuffix(domain, config.ToasterDomain) {
		utils.SendError(w, "custom domains are not yet supported", "invalidHost", 400)
		return
	}

	spl := strings.Split(domain, ".")

	if len(spl) <= len(config.ToasterDomainSplitted) {
		utils.SendError(w, "the requested host does not contain a subdomain or a toasterid", "invalidHost", 400)
		return
	}

	var toasteridOrSubdomain, requestedRegion string
	switch {
	case len(spl) == len(config.ToasterDomainSplitted)+1:
		toasteridOrSubdomain = spl[0]
	case len(spl) == len(config.ToasterDomainSplitted)+2:
		toasteridOrSubdomain = spl[0]
		requestedRegion = spl[1]
	default:
		utils.SendError(w, "the requested host contains an invalid number of subdomains", "invalidHost", 400)
		return
	}

	exeid = r.Header.Get("X-TOASTAINER-EXEID")

	if exeid == "" {
		if strings.HasPrefix(r.URL.Path, "/ex_") {
			spl := strings.Split(r.URL.Path, "/")
			exeid = spl[1]

			if len(spl) > 2 {
				r.URL.Path = "/" + strings.Join(spl[2:], "/")
			} else {
				r.URL.Path = "/"
			}
		}

		if exeid == "" {
			tmp, ok := r.URL.Query()["exeid"]

			if ok && len(tmp) > 0 {
				exeid = tmp[0]
			}
		}
	}

	toaster.RunToaster(w, r, exeid, toasteridOrSubdomain, requestedRegion, s.isHTTPS)
}

func setupCORS(w http.ResponseWriter, origin string) {
	h := w.Header()

	// allow-Origin as wildcard and allow credentials are not allowed both at the same time: https://portswigger.net/web-security/cors/access-control-allow-origin
	if origin == "" {
		h.Add("Access-Control-Allow-Origin", "*")
	} else {
		h.Add("Access-Control-Allow-Origin", origin)
		h.Add("Access-Control-Allow-Credentials", "true")
	}

	h.Add("Access-Control-Allow-Methods", "POST, PUT, GET, DELETE, OPTIONS")
	h.Add("Access-Control-Allow-Headers", "X-TOASTAINER-AUTH,X-TOASTAINER-EXEID,X-TOASTAINER-EXE-TIMEOUT-SEC,Origin,Accept,Access-Control-Allow-Origin,Access-Control-Allow-Methods,Access-Control-Allow-Headers,Access-Control-Allow-Credentials,Accept-Encoding,Accept-Language,Access-Control-Request-Headers,Access-Control-Request-Method,Cache-Control,Connection,Host,Pragma,Referer,Sec-Fetch-Dest,Sec-Fetch-Mode,Sec-Fetch-Site,Set-Cookie,User-Agent,Vary,Method,Content-Type,Content-Length") // todo: étoffer
	h.Add("Vary", "*")
	h.Add("Cache-Control", "no-store")
}
