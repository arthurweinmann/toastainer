package acme

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/providers/dns"
	"github.com/go-acme/lego/v4/registration"
	"github.com/toastate/toastcloud/internal/config"
	"github.com/toastate/toastcloud/internal/db/objectdb"
	"github.com/toastate/toastcloud/internal/db/objectdb/objectdberror"
	"github.com/toastate/toastcloud/internal/db/redisdb"
	"github.com/toastate/toastcloud/internal/model"
	"github.com/toastate/toastcloud/internal/nodes"
	"github.com/toastate/toastcloud/internal/utils"
)

var ErrCertificateNotFound = errors.New("certificate not found")
var ErrCertificateExpired = errors.New("certificate expired")

var legoconfig *lego.Config
var client *lego.Client
var reg *registration.Resource
var createMu sync.RWMutex
var httpChal *HTTPChallenger

func createHandler(us *ACMEUser, isnew bool) error {
	var err error

	legoconfig = lego.NewConfig(us)
	legoconfig.CADirURL = lego.LEDirectoryProduction
	// lego.LEDirectoryStaging

	client, err = lego.NewClient(legoconfig)
	if err != nil {
		return err
	}

	httpChal = &HTTPChallenger{}
	err = client.Challenge.SetHTTP01Provider(httpChal)
	if err != nil {
		return err
	}

	if config.DNSProvider.Name != "" {
		cp, err := dns.NewDNSChallengeProviderByName(config.DNSProvider.Name)
		if err != nil {
			return err
		}
		err = client.Challenge.SetDNS01Provider(cp)
		if err != nil {
			return err
		}
	}

	if isnew {
		// New users will need to register
		reg, err = client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
		if err != nil {
			return err
		}
		us.Registration = reg

		err = us.SaveOnDisk()
		if err != nil {
			return err
		}
	} else {
		// check registration
		reg, err = client.Registration.QueryRegistration()
		if err != nil {
			return err
		}
	}

	return nil
}

func CreateCertificate(rootdomain string, domains []string, lock bool) ([]byte, []byte, error) {
	var certificates *certificate.Resource
	var err error

	if lock {
		ok, err := LockCert(rootdomain)
		if err != nil {
			return nil, nil, err
		}
		if !ok {
			return nil, nil, nil
		}
		defer UnlockCert(rootdomain)
	}

	certificates, err = client.Certificate.Obtain(certificate.ObtainRequest{
		Domains: domains,
		Bundle:  true,
	})
	if err != nil {
		fmt.Println("Obtain cert err", rootdomain, err)
		return nil, nil, err
	}

	err = storeCertificate(rootdomain, certificates.Certificate, certificates.PrivateKey)
	if err != nil {
		return nil, nil, err
	}

	return certificates.Certificate, certificates.PrivateKey, nil
}

func LockCert(rootdomain string) (bool, error) {
	ok, err := redisdb.LockTryOnce("___certLock_"+rootdomain, 5*time.Minute)
	if err != nil {
		return false, err
	}
	return ok, nil
}

func UnlockCert(rootdomain string) error {
	return redisdb.Unlock("___certLock_" + rootdomain)
}

func getCertificate(domain string) (*tls.Certificate, error) {
	cert, priv, err := RetrieveCertificate(domain)
	if err != nil {
		switch err {
		case ErrCertificateNotFound, ErrCertificateExpired:
		default:
			return nil, fmt.Errorf("RetrieveCertificate: %v", err)
		}

		rootdomain, err := utils.ExtractRootDomain(domain)
		if err != nil {
			return nil, err
		}

		// ToasterRootDomain and APIRootDomain may be the same domain
		// but best practice is to make them distinct for cookie redundant security purposes
		var domains []string
		switch rootdomain {
		case config.ToasterRootDomain:
			domains = append([]string{"*." + config.ToasterDomain}, nodes.GetToasterLocalRegionAppSubdomain(config.ToasterDomain, config.Region))
			if config.ToasterRootDomain == config.APIRootDomain {
				domains = append(domains, config.APIDomain)
			}

		case config.APIRootDomain:
			domains = []string{config.APIDomain}
			if config.ToasterRootDomain == config.APIRootDomain {
				domains = append(domains, "*."+config.ToasterDomain)
				domains = append(domains, nodes.GetToasterLocalRegionAppSubdomain(config.ToasterDomain, config.Region))
			}

		default:
			return nil, fmt.Errorf("custom domains are not yet supported")
		}

		cert, priv, err = CreateCertificate(rootdomain, domains, true)
		if err != nil {
			return nil, err
		}
		if cert == nil {
			return nil, nil
		}
	}

	return GenerateCert(cert, priv)
}

func ToggleCertificate(domains []string) error {
	rootdomain, err := utils.ExtractRootDomain(domains[0])
	if err != nil {
		return err
	}

	_, _, err = RetrieveCertificate(rootdomain)
	if err != nil {
		if err != ErrCertificateExpired && err != ErrCertificateNotFound {
			return err
		}

		_, _, err = CreateCertificate(rootdomain, domains, true)
		if err != nil {
			return err
		}
	}

	return nil
}

// TODO: store the list of subdomains too in order to recreate the cert if this list has changed
func storeCertificate(domain string, certificate, privateKey []byte) error {
	b := make([]byte, 10, len(certificate)+len(privateKey)+10)

	binary.BigEndian.PutUint64(b[:8], uint64(time.Now().Add(744*time.Hour).Unix()))
	binary.BigEndian.PutUint16(b[8:], uint16(len(certificate)))

	b = append(b, certificate...)
	b = append(b, privateKey...)

	err := objectdb.Client.UpsertCertificate(&model.Certificate{
		Domain: domain,
		Cert:   b,
	})
	if err != nil {
		return err
	}

	cache.Set(utils.String2ByteSlice(domain), b)

	err = redisdb.GetClient().Set(context.Background(), "cert_"+domain, b, 744*2*time.Hour).Err()
	if err != nil {
		return err
	}

	return nil
}

func RetrieveCertificate(domain string) (certificate, privateKey []byte, err error) {
	b := cache.Get(nil, utils.String2ByteSlice(domain))

	var updateRedisAntiSnowfall bool
	if len(b) == 0 {
		b, err = redisdb.GetClient().Get(context.Background(), "cert_"+domain).Bytes()
		if err == redisdb.ErrNil {
			err = ErrCertificateNotFound
			updateRedisAntiSnowfall = true
		}

		if err != nil && len(b) > 0 {
			cache.Set(utils.String2ByteSlice(domain), b)
		}
	}

	if len(b) == 0 {
		cert, err := objectdb.Client.GetCertificate(domain)
		if err != nil {
			if err == objectdberror.ErrNotFound {
				err = ErrCertificateNotFound
			}
		} else {
			b = cert.Cert
		}

		if err != nil && len(b) > 0 && updateRedisAntiSnowfall {
			redisdb.GetClient().Set(context.Background(), "cert_"+domain, b, 744*2*time.Hour)
		}
	}

	if len(b) == 0 && err == nil {
		err = ErrCertificateNotFound
	}

	if err != nil {
		return
	}

	if uint64(time.Now().Unix()) > binary.BigEndian.Uint64(b[:8]) {
		return nil, nil, ErrCertificateExpired
	}

	lc := binary.BigEndian.Uint16(b[8:])
	certificate = b[10 : lc+10]
	privateKey = b[lc+10:]

	return
}

func GenerateCert(certificate []byte, privateKey []byte) (*tls.Certificate, error) {
	// Leaf is nil when using this method, see if we need to provide id
	// See https://stackoverflow.com/questions/43605755/whats-the-leaf-certificate-and-sub-certificate-used-for-and-how-to-use-them
	cert, err := tls.X509KeyPair(certificate, privateKey)
	if err != nil {
		return nil, err
	}
	return &cert, nil
}
