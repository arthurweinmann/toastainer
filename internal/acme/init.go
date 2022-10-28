package acme

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"net/mail"

	"github.com/VictoriaMetrics/fastcache"
	"github.com/toastate/toastainer/internal/config"
	"github.com/toastate/toastainer/internal/nodes"
)

var cache *fastcache.Cache

func Init() error {
	cache = fastcache.New(32 * 1024 * 1024)

	if config.CertificateContactEmail == "" {
		return fmt.Errorf("you must provide a certificate contact email in the configuration")
	}

	e, err := mail.ParseAddress(config.CertificateContactEmail)
	if err != nil {
		return fmt.Errorf("invalid certificate contact email address: %v", err)
	}
	config.CertificateContactEmail = e.Address

	us, err := loadACMEUserFromDisk()
	if err != nil {
		return err
	}

	if us == nil {
		// Create a user. New accounts need an email and private key to start.
		privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return err
		}

		us = &ACMEUser{
			Email: config.CertificateContactEmail,
			key:   privateKey,
		}

		err = createHandler(us, true)
		if err != nil {
			return err
		}

		return nil
	}

	err = createHandler(us, false)
	if err != nil {
		return err
	}

	builtins := BuiltinCerts()
	for i := 0; i < len(builtins); i++ {
		err := ToggleCertificate(builtins[i])
		if err != nil {
			return err
		}
	}

	return nil
}

func BuiltinCerts() [][]string {
	perRootDomains := map[string][]string{}

	perRootDomains[config.ToasterRootDomain] = append(perRootDomains[config.ToasterRootDomain], "*."+config.ToasterDomain)
	if config.Region != "" {
		perRootDomains[config.ToasterRootDomain] = append(perRootDomains[config.ToasterRootDomain], nodes.GetToasterLocalRegionAppSubdomain(config.ToasterDomain, config.Region))
	}

	perRootDomains[config.APIRootDomain] = append(perRootDomains[config.APIRootDomain], config.APIDomain)

	perRootDomains[config.DashboardRootDomain] = append(perRootDomains[config.DashboardRootDomain], config.DashboardDomain)

	var ret [][]string

	for _, v := range perRootDomains {
		ret = append(ret, v)
	}

	return ret
}
