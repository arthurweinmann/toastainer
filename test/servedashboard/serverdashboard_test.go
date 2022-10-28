package servedashboard

import (
	"crypto/tls"
	"crypto/x509/pkix"
	"fmt"
	"log"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/toastate/toastainer/internal/acme"
	"github.com/toastate/toastainer/internal/api"
	"github.com/toastate/toastainer/internal/config"
	"github.com/toastate/toastainer/internal/supervisor"
	"github.com/toastate/toastainer/internal/utils"
	"github.com/toastate/toastainer/test/helpers"

	_ "embed"
)

// TODO: certificate authority for browsers to open the website

//go:embed config_test.json
var config_test []byte

// You should use go test -timeout 99999s to run this test dashboard
func TestServeDashboard(t *testing.T) {
	err := serveDashboard(t)
	if err != nil {
		log.Fatal(fmt.Errorf("TestServeDashboard.serveDashboard: %v", err))
	}
}

func serveDashboard(t *testing.T) error {
	err := config.LoadConfigBytes(config_test, "json")
	if err != nil {
		return err
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	config.Home = filepath.Join(wd, "../../build")

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
	utils.InitLogging(config.LogLevel)

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

	var ca *utils.CertificateAuthority
	if utils.DirExists(filepath.Join(config.Home, "localCA")) && !utils.DirEmpty(filepath.Join(config.Home, "localCA")) {
		b, err := os.ReadFile(filepath.Join(config.Home, "localCA/certificateauthority.marshaled"))
		if err != nil {
			return err
		}

		ca, err = utils.UnmarshalCertificateAuthority(b)
		if err != nil {
			return err
		}
	} else {
		ca, err = utils.NewCertificateAuthority(pkix.Name{
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

		err = os.MkdirAll(filepath.Join(config.Home, "localCA"), 0700)
		if err != nil {
			return err
		}

		b, err := ca.Marshal()
		if err != nil {
			return err
		}

		err = os.WriteFile(filepath.Join(config.Home, "localCA/certificateauthority.marshaled"), b, 0644)
		if err != nil {
			return err
		}

		err = os.WriteFile(filepath.Join(config.Home, "localCA/root_certificate_load_in_browser"), ca.CertificateBytes(), 0644)
		if err != nil {
			return err
		}

		err = os.Chown(filepath.Join(config.Home, "localCA"), config.Runner.NonRootUID, config.Runner.NonRootGID)
		if err != nil {
			return err
		}

		err = os.Chown(filepath.Join(config.Home, "localCA/certificateauthority.marshaled"), config.Runner.NonRootUID, config.Runner.NonRootGID)
		if err != nil {
			return err
		}

		err = os.Chown(filepath.Join(config.Home, "localCA/root_certificate_load_in_browser"), config.Runner.NonRootUID, config.Runner.NonRootGID)
		if err != nil {
			return err
		}
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

	_, port := utils.SplitHostPort(strings.Replace(ts.URL, "https://", "", 1))
	utils.Info("msg", "Started HTTPS Test Server on https://"+config.DashboardDomain+":"+port)

	hostredicter := helpers.NewETChostmodifier()
	defer hostredicter.Reset()

	err = hostredicter.SetHostRedirection("127.0.0.1", config.DashboardDomain)
	if err != nil {
		return err
	}

	err = hostredicter.SetHostRedirection("127.0.0.1", config.APIDomain)
	if err != nil {
		return err
	}

	cmd := exec.Command("/bin/bash", "-c", "rm -rf ../build/web && mkdir ../build/web && TOASTAINER_API_HOST="+fmt.Sprintf("%q", config.APIDomain+":"+port)+" ../build/toastfront serve --build-dir=../build/web")
	cmd.Dir = filepath.Join(wd, "../../web")
	cmd.CombinedOutput()

	wat.BlockUntilShutdownDone()

	return nil
}
