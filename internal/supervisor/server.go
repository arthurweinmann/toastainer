package supervisor

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/toastate/toastainer/internal/acme"
	"github.com/toastate/toastainer/internal/api"
	"github.com/toastate/toastainer/internal/config"
	"github.com/toastate/toastainer/internal/utils"
)

type httpservers struct {
	errs  chan error
	http  *http.Server
	https *http.Server
}

func startServer() (s *httpservers, err error) {
	s = &httpservers{
		errs: make(chan error, 2),
	}

	defer func() {
		if err != nil {
			s.Close()
		}
	}()

	go s.startHTTP()
	time.Sleep(3 * time.Second)

	select {
	case err = <-s.errs:
		return
	default:
	}

	err = acme.Init()
	if err != nil {
		return
	}

	go s.startHTTPS()
	time.Sleep(3 * time.Second)

	select {
	case err = <-s.errs:
	default:
	}

	return
}

func (s *httpservers) startHTTPS() {
	var err error

	conn, err := net.Listen("tcp", ":443")
	if err != nil {
		s.errs <- fmt.Errorf("Couldn't bind to TCP socket %q: %s", ":443", err)
		return
	}
	tlsConfig := new(tls.Config)
	tlsConfig.GetCertificate = acme.GetCertificate
	tlsListener := tls.NewListener(conn, tlsConfig)

	if utils.FileExists(filepath.Join(config.Home, "https.log")) {
		err = os.Remove(filepath.Join(config.Home, "https.log"))
		if err != nil {
			s.errs <- fmt.Errorf("Couldn't remove %s: %v", filepath.Join(config.Home, "https.log"), err)
			return
		}
	}

	f, err := os.OpenFile(filepath.Join(config.Home, "https.log"), os.O_CREATE|os.O_RDWR, 0700)
	if err != nil {
		s.errs <- fmt.Errorf("Couldn't open file %s: %v", filepath.Join(config.Home, "https.log"), err)
		return
	}
	defer f.Close()

	s.https = &http.Server{
		Addr:     ":443",
		Handler:  api.NewHTTPSRouter(),
		ErrorLog: log.New(f, "https: ", log.Llongfile|log.Ltime|log.Ldate),

		ReadHeaderTimeout: 30 * time.Second,
		ReadTimeout:       1 * time.Minute,
		WriteTimeout:      1 * time.Minute,
		IdleTimeout:       5 * time.Minute,
	}

	s.errs <- s.https.Serve(tlsListener)
}

func (s *httpservers) startHTTP() {
	s.http = &http.Server{
		Addr:    ":80",
		Handler: api.NewHTTPRouter(),

		ReadHeaderTimeout: 30 * time.Second,
		ReadTimeout:       1 * time.Minute,
		WriteTimeout:      1 * time.Minute,
		IdleTimeout:       5 * time.Minute,
	}

	s.errs <- s.http.ListenAndServe()
}

func (s *httpservers) Close() {
	if s.https != nil {
		s.https.Close()
	}

	if s.http != nil {
		s.http.Close()
	}
}
