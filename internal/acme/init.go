package acme

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"net/mail"

	"github.com/VictoriaMetrics/fastcache"
	"github.com/toastate/toastcloud/internal/config"
	"github.com/toastate/toastcloud/internal/nodes"
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

	if config.APIRootDomain == config.ToasterRootDomain {
		d := append([]string{config.APIDomain, "*." + config.ToasterDomain}, nodes.GetToasterLocalRegionAppSubdomain(config.ToasterDomain, config.Region))
		err := ToggleCertificate(d)
		if err != nil {
			return err
		}
	} else {
		d := append([]string{"*." + config.ToasterDomain}, nodes.GetToasterLocalRegionAppSubdomain(config.ToasterDomain, config.Region))

		err := ToggleCertificate(d)
		if err != nil {
			return err
		}

		err = ToggleCertificate([]string{config.APIDomain})
		if err != nil {
			return err
		}
	}

	return nil
}
