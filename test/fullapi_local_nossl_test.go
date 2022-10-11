package test

import (
	"bytes"
	"crypto/tls"
	"crypto/x509/pkix"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/toastate/toastainer/internal/acme"
	"github.com/toastate/toastainer/internal/api"
	"github.com/toastate/toastainer/internal/config"
	"github.com/toastate/toastainer/internal/supervisor"
	"github.com/toastate/toastainer/internal/utils"
	"github.com/toastate/toastainer/test/library"

	_ "embed"
)

//go:embed config_test.json
var config_test []byte

func TestFullAPILocalNoSSL(t *testing.T) {
	err := fullAPILocalNoSSL()
	if err != nil {
		log.Fatal(fmt.Errorf("TestFullAPILocalNoSSL.fullAPILocalNoSSL: %v", err))
	}
}

func fullAPILocalNoSSL() error {
	err := config.LoadConfigBytes(config_test, "json")
	if err != nil {
		return err
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	config.Home = filepath.Join(wd, "../build")

	tmpdir, err := os.MkdirTemp(config.Home, "fullapitest")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpdir)

	err = os.Chown(tmpdir, config.Runner.NonRootUID, config.Runner.NonRootGID)
	if err != nil {
		return err
	}

	config.Runner.BTRFSFile = filepath.Join(tmpdir, "btrfsfile")
	config.Runner.BTRFSMountPoint = filepath.Join(tmpdir, "btrfs")
	config.Runner.OverlayMountPoint = filepath.Join(tmpdir, "overlay")
	config.ObjectStorage.LocalFS.Path = filepath.Join(tmpdir, "objectstorage")

	if utils.DirEmpty(config.Home) {
		return fmt.Errorf("empty or inacessible project build directory")
	}

	config.LogLevel = "all"
	utils.InitLogging()

	if utils.FileExists(filepath.Join(config.Home, "nsjail.log")) {
		err := os.Remove(filepath.Join(config.Home, "nsjail.log"))
		if err != nil {
			return err
		}
	}

	wat, err := supervisor.StartNoServer()
	if err != nil {
		return fmt.Errorf("supervisor.StartNoServer: %v", err)
	}
	defer wat.Shutdown()

	ts := httptest.NewUnstartedServer(api.NewHTTPSRouter())
	ts.TLS = new(tls.Config)

	ca, err := utils.NewCertificateAuthority(pkix.Name{
		Organization:  []string{"TOASTATE.COM SAS"},
		Country:       []string{"FR"},
		Province:      []string{""},
		Locality:      []string{"Strasbourg"},
		StreetAddress: []string{""},
		PostalCode:    []string{"67000"},
	})
	if err != nil {
		return err
	}

	bc := acme.BuiltinCerts()
	for i := 1; i < len(bc); i++ {
		bc[0] = append(bc[0], bc[i]...)
	}

	cert, err := ca.CreateDNSRSACertificate(bc[0])
	if err != nil {
		return err
	}

	ts.TLS.Certificates = []tls.Certificate{cert}

	ts.StartTLS()
	defer ts.Close()

	utils.Info("msg", "Started HTTPS Test Server on "+ts.URL)

	hostredicter := newetchostmodifier()
	defer hostredicter.Reset()

	fat, err := library.NewFullAPITest(func() *http.Client {
		c := &http.Client{Transport: ts.Client().Transport}
		return c
	}, ts.URL, config.APIDomain, config.ToasterDomain, &library.FullAPITestOpts{
		SetHostRedirection: hostredicter.SetHostRedirection,
	})
	if err != nil {
		return fmt.Errorf("could not setup full api test: %v", err)
	}

	err = fat.Run()
	if err != nil {
		return fmt.Errorf("full api test error: %v", err)
	}

	log.Println("All tests passed successfully")
	return nil
}

type etchostmodifier struct {
	original []byte
	mu       sync.Mutex
}

func newetchostmodifier() *etchostmodifier {
	return &etchostmodifier{}
}

func (m *etchostmodifier) SetHostRedirection(ip, hostname string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	b, err := os.ReadFile("/etc/hosts")
	if err != nil {
		return err
	}

	if m.original == nil {
		m.original = b
	}

	line := ip + " " + hostname

	if !bytes.Contains(b, []byte(line)) {
		b = append(b, fmt.Sprintf("\n%v\n", line)...)
		err = os.Truncate("/etc/hosts", 0)
		if err != nil {
			return err
		}
		err = os.WriteFile("/etc/hosts", b, 0777)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *etchostmodifier) Reset() error {
	if m.original != nil {
		err := os.Truncate("/etc/hosts", 0)
		if err != nil {
			return err
		}
		err = os.WriteFile("/etc/hosts", m.original, 0644)
		if err != nil {
			return err
		}
	}

	return nil
}
