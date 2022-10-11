package sqldb

import (
	"database/sql"

	"github.com/toastate/toastainer/internal/db/objectdb/objectdberror"
	"github.com/toastate/toastainer/internal/model"
)

func (c *Client) UpsertCertificate(cert *model.Certificate) error {
	_, err := c.db.Exec("INSERT INTO certificates(domain, cert) VALUES (?,?) ON DUPLICATE KEY UPDATE cert = ?", cert.Domain, cert.Cert, cert.Cert)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) GetCertificate(domain string) (*model.Certificate, error) {
	cert := []model.Certificate{}
	err := c.db.Select(&cert, "SELECT * FROM certificates WHERE domain = ?", domain)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, objectdberror.ErrNotFound
		}
		return nil, err
	}
	if len(cert) == 0 {
		return nil, objectdberror.ErrNotFound
	}

	return &cert[0], nil
}

func (c *Client) DelCertificate(domain string) error {
	_, err := c.db.Exec("DELETE FROM certificates WHERE domain = ?", domain)
	if err != nil {
		if err == sql.ErrNoRows {
			return objectdberror.ErrNotFound
		}
		return err
	}

	return nil
}
