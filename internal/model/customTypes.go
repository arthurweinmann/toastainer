package model

import (
	"database/sql/driver"
	"encoding/binary"
	"errors"
)

type ArrayString []string

// Value implements the driver.Valuer interface, returning a []byte
func (g ArrayString) Value() (driver.Value, error) {
	b := make([]byte, 0, len(g)*4)

	for i := 0; i < len(g); i++ {
		b = append(b, 0, 0)
		binary.BigEndian.PutUint16(b[len(b)-2:], uint16(len(g[i])))
		b = append(b, g[i]...)
	}

	return b, nil
}

// Scan implements the sql.Scanner interface
func (g *ArrayString) Scan(src interface{}) error {
	var source []byte
	switch src := src.(type) {
	case string:
		source = []byte(src)
	case []byte:
		source = src
	default:
		return errors.New("Incompatible type for ArrayString")
	}

	var arrstr []string
	i := 0
	var l int
	for i < len(source) {
		l = int(binary.BigEndian.Uint16(source[i : i+2]))
		i += 2

		if i+int(l) >= len(source) {
			return errors.New("received invalid marshaled ArrayString from database")
		}

		arrstr = append(arrstr, string(source[i:i+l]))
		i += l
	}

	*g = arrstr
	return nil
}
