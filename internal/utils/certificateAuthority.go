package utils

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/binary"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"time"
)

type CertificateAuthority struct {
	root                *x509.Certificate
	privKey             *rsa.PrivateKey
	certificatePEMBlock []byte
	privateKeyPEMBlock  []byte
	organization        pkix.Name
}

func NewCertificateAuthority(organization pkix.Name) (*CertificateAuthority, error) {
	var err error

	ca := &CertificateAuthority{
		organization: organization,
	}

	ca.root = &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               organization,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	ca.privKey, err = rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, err
	}

	caBytes, err := x509.CreateCertificate(rand.Reader, ca.root, ca.root, &ca.privKey.PublicKey, ca.privKey)
	if err != nil {
		return nil, err
	}

	caPEM := new(bytes.Buffer)
	err = pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})
	if err != nil {
		return nil, err
	}

	ca.certificatePEMBlock = caPEM.Bytes()

	caPrivKeyPEM := new(bytes.Buffer)
	err = pem.Encode(caPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(ca.privKey),
	})
	if err != nil {
		return nil, err
	}

	ca.privateKeyPEMBlock = caPrivKeyPEM.Bytes()

	return ca, nil
}

func (ca *CertificateAuthority) CreateDNSRSACertificate(dnsNames []string) (tls.Certificate, error) {
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject:      ca.organization,
		DNSNames:     dnsNames,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6}, // see https://ldapwiki.com/wiki/SubjectKeyIdentifier
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	certPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return tls.Certificate{}, err
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, ca.root, &certPrivKey.PublicKey, ca.privKey)
	if err != nil {
		return tls.Certificate{}, err
	}

	certPEM := new(bytes.Buffer)
	pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	certPrivKeyPEM := new(bytes.Buffer)
	pem.Encode(certPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certPrivKey),
	})

	return tls.X509KeyPair(certPEM.Bytes(), certPrivKeyPEM.Bytes())
}

func (ca *CertificateAuthority) CertificateBytes() []byte {
	return ca.certificatePEMBlock
}

func (ca *CertificateAuthority) Marshal() ([]byte, error) {
	ret := make([]byte, 0)

	b, err := json.Marshal(ca.organization)
	if err != nil {
		return nil, err
	}
	ret = binary.BigEndian.AppendUint64(ret, uint64(len(b)))
	ret = append(ret, b...)

	ret = binary.BigEndian.AppendUint64(ret, uint64(len(ca.certificatePEMBlock)))
	ret = append(ret, ca.certificatePEMBlock...)

	ret = binary.BigEndian.AppendUint64(ret, uint64(len(ca.privateKeyPEMBlock)))
	ret = append(ret, ca.privateKeyPEMBlock...)

	var caPrivKeybuf bytes.Buffer
	by := x509.MarshalPKCS1PrivateKey(ca.privKey)
	pb := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: by}
	err = pem.Encode(&caPrivKeybuf, pb)
	if err != nil {
		return nil, err
	}
	marshaledCAprivKey := caPrivKeybuf.Bytes()
	ret = binary.BigEndian.AppendUint64(ret, uint64(len(marshaledCAprivKey)))
	ret = append(ret, marshaledCAprivKey...)
	by, err = ca.root.SerialNumber.MarshalText()
	if err != nil {
		return nil, err
	}
	ret = binary.BigEndian.AppendUint64(ret, uint64(len(by)))
	ret = append(ret, by...)
	by, err = json.Marshal(ca.root.Subject)
	if err != nil {
		return nil, err
	}
	ret = binary.BigEndian.AppendUint64(ret, uint64(len(by)))
	ret = append(ret, by...)
	ret = binary.BigEndian.AppendUint64(ret, uint64(ca.root.NotBefore.UnixNano()))
	ret = binary.BigEndian.AppendUint64(ret, uint64(ca.root.NotAfter.UnixNano()))

	return ret, nil
}

func UnmarshalCertificateAuthority(b []byte) (*CertificateAuthority, error) {
	ret := &CertificateAuthority{
		root: &x509.Certificate{
			IsCA:                  true,
			ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
			KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
			BasicConstraintsValid: true,
		},
	}

	offset := uint64(0)
	l := binary.BigEndian.Uint64(b[offset : offset+8])
	offset += 8
	err := json.Unmarshal(b[offset:offset+l], &ret.organization)
	offset += l
	if err != nil {
		return nil, err
	}

	l = binary.BigEndian.Uint64(b[offset : offset+8])
	offset += 8
	ret.certificatePEMBlock = b[offset : offset+l]
	offset += l

	l = binary.BigEndian.Uint64(b[offset : offset+8])
	offset += 8
	ret.privateKeyPEMBlock = b[offset : offset+l]
	offset += l

	l = binary.BigEndian.Uint64(b[offset : offset+8])
	offset += 8
	p, _ := pem.Decode(b[offset : offset+l])
	offset += l
	if p == nil {
		return nil, fmt.Errorf("could not decode ca private key pem block bytes")
	}
	privKey, err := parsePrivateKey(p.Bytes)
	if err != nil {
		return nil, err
	}
	ret.privKey, _ = privKey.(*rsa.PrivateKey)
	if ret.privKey == nil {
		return nil, fmt.Errorf("could not decode ca private key pem block bytes")
	}

	l = binary.BigEndian.Uint64(b[offset : offset+8])
	offset += 8
	bi := big.NewInt(0)
	err = bi.UnmarshalText(b[offset : offset+l])
	offset += l
	if err != nil {
		return nil, err
	}
	ret.root.SerialNumber = bi

	l = binary.BigEndian.Uint64(b[offset : offset+8])
	offset += 8
	err = json.Unmarshal(b[offset:offset+l], &ret.root.Subject)
	offset += l
	if err != nil {
		return nil, err
	}

	l = binary.BigEndian.Uint64(b[offset : offset+8])
	offset += 8
	ret.root.NotBefore = time.Unix(0, int64(l))

	l = binary.BigEndian.Uint64(b[offset : offset+8])
	offset += 8
	ret.root.NotAfter = time.Unix(0, int64(l))

	return ret, nil
}

// From parsePrivateKey in crypto/tls/tls.go.
func parsePrivateKey(der []byte) (crypto.Signer, error) {
	if key, err := x509.ParsePKCS1PrivateKey(der); err == nil {
		return key, nil
	}
	if key, err := x509.ParsePKCS8PrivateKey(der); err == nil {
		switch key := key.(type) {
		case *rsa.PrivateKey:
			return key, nil
		case *ecdsa.PrivateKey:
			return key, nil
		default:
			return nil, errors.New("unknown private key type in PKCS#8 wrapping")
		}
	}
	if key, err := x509.ParseECPrivateKey(der); err == nil {
		return key, nil
	}

	return nil, errors.New("failed to parse private key")
}
