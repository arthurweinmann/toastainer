package acme

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/go-acme/lego/v4/registration"
	"github.com/toastate/toastcloud/internal/config"
	"github.com/toastate/toastcloud/internal/utils"
)

type ACMEUser struct {
	Email        string                 `json:"email"`
	Registration *registration.Resource `json:"registration"`
	Key          string                 `json:"key"`

	key *ecdsa.PrivateKey
}

func (u *ACMEUser) GetEmail() string {
	return u.Email
}

func (u ACMEUser) GetRegistration() *registration.Resource {
	return u.Registration
}

func (u *ACMEUser) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

func (u *ACMEUser) SaveOnDisk() error {
	p := filepath.Join(config.Home, "acme/account.json")

	if utils.FileExists(p) {
		err := os.Remove(p)
		if err != nil {
			return err
		}
	}

	u.Key = encode(u.key)
	defer func() {
		u.Key = ""
	}()

	b, err := json.Marshal(u)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(p, b, 0700)
}

func loadACMEUserFromDisk() (*ACMEUser, error) {
	p := filepath.Join(config.Home, "acme")
	err := os.MkdirAll(p, 0755)
	if err != nil {
		return nil, err
	}

	p = filepath.Join(config.Home, "acme/account.json")

	if !utils.FileExists(p) {
		return nil, nil
	}

	b, err := ioutil.ReadFile(p)
	if err != nil {
		return nil, err
	}

	u := &ACMEUser{}
	err = json.Unmarshal(b, u)
	if err != nil {
		return nil, err
	}

	u.key = decode(u.Key)
	u.Key = ""

	return u, nil
}

func encode(privateKey *ecdsa.PrivateKey) string {
	x509Encoded, _ := x509.MarshalECPrivateKey(privateKey)
	pemEncoded := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: x509Encoded})

	return string(pemEncoded)
}

func decode(pemEncoded string) *ecdsa.PrivateKey {
	block, _ := pem.Decode([]byte(pemEncoded))
	x509Encoded := block.Bytes
	privateKey, _ := x509.ParseECPrivateKey(x509Encoded)

	return privateKey
}
