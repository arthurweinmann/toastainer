package sqldb

import (
	"database/sql"
	"fmt"
	"sort"
	"strconv"

	"github.com/toastate/toastainer/internal/db/objectdb/objectdberror"
	"github.com/toastate/toastainer/internal/model"
)

func (c *Client) CreateUser(usr *model.User) error {
	_, err := c.db.Exec("INSERT INTO users(id, email, password) VALUES (?,?,?)", usr.ID, usr.Email, usr.Password)
	if err != nil {
		// We need to query if the user already exists because error codes vary for each sql backends
		// except for ErrNoRows
		ok, _ := c.UserExistsByEmail(usr.Email)
		if ok {
			return objectdberror.ErrAlreadyExists
		}

		return err
	}

	return nil
}

func (c *Client) UpdateUser(usr *model.User) error {
	_, err := c.db.Exec("UPDATE users SET email = ?, password = ? WHERE id = ?", usr.Email, usr.Password, usr.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return objectdberror.ErrNotFound
		}
		return err
	}

	return nil
}

func (c *Client) GetUserByEmail(email string) (*model.User, error) {
	usrs := []model.User{}
	err := c.db.Select(&usrs, "SELECT * FROM users WHERE email=?", email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, objectdberror.ErrNotFound
		}
		return nil, err
	}
	if len(usrs) == 0 {
		return nil, objectdberror.ErrNotFound
	}

	return &usrs[0], nil
}

func (c *Client) GetUserByID(userid string) (*model.User, error) {
	usrs := []model.User{}
	err := c.db.Select(&usrs, "SELECT * FROM users WHERE id=?", userid)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, objectdberror.ErrNotFound
		}
		return nil, err
	}
	if len(usrs) == 0 {
		return nil, objectdberror.ErrNotFound
	}

	return &usrs[0], nil
}

func (c *Client) RangeUsers(limit int, cursor string) (string, bool, []model.User, error) {
	var err error

	usrs := []model.User{}

	var nc int
	if cursor == "" {
		nc = 0
	} else {
		nc, err = strconv.Atoi(cursor)
		if err != nil {
			return "", false, nil, fmt.Errorf("invalid cursor")
		}
	}

	// The %d makes sure there can be no SQL injection, be careful when modifying this line
	req := fmt.Sprintf("SELECT * FROM (SELECT * FROM users WHERE cursor>%d) subq ORDER BY cursor ASC LIMIT %d", nc, limit+1)
	err = c.db.Select(&usrs, req)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", false, usrs, nil
		}
		return "", false, nil, err
	}

	sort.Slice(usrs, func(i, j int) bool {
		return usrs[i].Cursor < usrs[j].Cursor
	})

	cursor = strconv.Itoa(usrs[len(usrs)-1].Cursor)

	if len(usrs) > limit {
		return cursor, false, usrs[:limit], nil
	}

	return cursor, false, usrs, nil
}

func (c *Client) UserExistsByEmail(email string) (bool, error) {
	var exists bool
	err := c.db.QueryRow("SELECT EXISTS(SELECT * FROM users WHERE email=?)", email).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return exists, nil
}

func (c *Client) DelUser(userid string) error {
	_, err := c.db.Exec("DELETE FROM users WHERE id = ?", userid)
	if err != nil {
		if err == sql.ErrNoRows {
			return objectdberror.ErrNotFound
		}
		return err
	}

	return nil
}
