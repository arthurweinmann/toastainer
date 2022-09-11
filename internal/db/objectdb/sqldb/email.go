package sqldb

import (
	"database/sql"
	"time"

	"github.com/toastate/toastcloud/internal/db/objectdb/objectdberror"
)

func (c *Client) BlockEmail(email, data string) error {
	_, err := c.db.Exec("INSERT INTO email_blocklist(date, email, data) VALUES (?,?,?)", time.Now().Unix(), email, data)
	if err != nil {
		// We need to query if the user already exists because error codes vary for each sql backends
		// except for ErrNoRows
		ok, _ := c.IsEmailBlocked(email)
		if ok {
			return objectdberror.ErrAlreadyExists
		}

		return err
	}

	return nil
}

func (c *Client) IsEmailBlocked(email string) (bool, error) {
	var exists bool
	err := c.db.QueryRow("SELECT EXISTS(SELECT * FROM email_blocklist WHERE email = ?", email).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return exists, nil
}
