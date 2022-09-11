package sqldb

import (
	"database/sql"

	"github.com/toastate/toastcloud/internal/db/objectdb/objectdberror"
	"github.com/toastate/toastcloud/internal/model"
)

// CreateSubdomain will return ErrAlreadyExists for already attributed subdomains
func (c *Client) CreateSubDomain(sub *model.SubDomain) error {
	_, err := c.db.Exec("INSERT INTO subdomains(id, name, user_id, toaster_id) VALUES (?,?,?,?)", sub.ID, sub.Name, sub.UserID, sub.ToasterID)
	if err != nil {
		// We need to query if the user already exists because error codes vary for each sql backends
		// except for ErrNoRows
		ok, _ := c.SubdomainExists(sub.Name)
		if ok {
			return objectdberror.ErrAlreadyExists
		}

		return err
	}

	return nil
}

func (c *Client) UpdateSubDomain(sub *model.SubDomain) error {
	_, err := c.db.Exec("UPDATE subdomains SET toaster_id = ? WHERE id = ?", sub.ToasterID, sub.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return objectdberror.ErrNotFound
		}
		return err
	}

	return nil
}

func (c *Client) GetLinkedSubDomains(toasterid string) ([]*model.SubDomain, error) {
	subs := []*model.SubDomain{}
	err := c.db.Select(&subs, "SELECT * FROM subdomains WHERE toaster_id = ?", toasterid)
	if err != nil {
		if err == sql.ErrNoRows {
			return subs, nil
		}
		return nil, err
	}

	return subs, nil
}

func (c *Client) ListUserSubdomains(userid string) ([]*model.SubDomain, error) {
	subs := []*model.SubDomain{}
	err := c.db.Select(&subs, "SELECT * FROM subdomains WHERE user_id = ?", userid)
	if err != nil {
		if err == sql.ErrNoRows {
			return subs, nil
		}
		return nil, err
	}

	return subs, nil
}

func (c *Client) CheckSubdomainOwnership(subname, userid string) (bool, error) {
	var exists bool
	err := c.db.QueryRow("SELECT EXISTS(SELECT * FROM subdomains WHERE name = ? AND user_id = ?)", subname, userid).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return exists, nil
}

func (c *Client) SubdomainExists(name string) (bool, error) {
	var exists bool
	err := c.db.QueryRow("SELECT EXISTS(SELECT * FROM subdomains WHERE name=?)", name).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return exists, nil
}

func (c *Client) GetSubDomain(userid, subdomainid string) (*model.SubDomain, error) {
	subs := []model.SubDomain{}
	err := c.db.Select(&subs, "SELECT * FROM subdomains WHERE id = ? AND user_id = ?", subdomainid, userid)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, objectdberror.ErrNotFound
		}
		return nil, err
	}
	if len(subs) == 0 {
		return nil, objectdberror.ErrNotFound
	}

	return &subs[0], nil
}

func (c *Client) DeleteSubDomain(userid, subdomainid string) error {
	_, err := c.db.Exec("DELETE FROM subdomains WHERE id = ? AND user_id = ?", subdomainid, userid)
	if err != nil {
		if err == sql.ErrNoRows {
			return objectdberror.ErrNotFound
		}
		return err
	}

	return nil
}

func (c *Client) DeleteToasterAllSubdomains(userid, toasterid string) error {
	_, err := c.db.Exec("DELETE FROM subdomains WHERE toaster_id = ? AND user_id = ?", toasterid, userid)
	if err != nil {
		if err == sql.ErrNoRows {
			return objectdberror.ErrNotFound
		}
		return err
	}

	return nil
}
