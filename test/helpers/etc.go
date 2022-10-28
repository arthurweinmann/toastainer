package helpers

import (
	"bytes"
	"fmt"
	"os"
	"sync"
)

type ETChostmodifier struct {
	original []byte
	mu       sync.Mutex
}

func NewETChostmodifier() *ETChostmodifier {
	return &ETChostmodifier{}
}

func (m *ETChostmodifier) SetHostRedirection(ip, hostname string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	b, err := os.ReadFile("/etc/hosts")
	if err != nil {
		return err
	}

	if m.original == nil {
		m.original = b
	}

	line := ip + " " + hostname

	if !bytes.Contains(b, []byte(line)) {
		b = append(b, fmt.Sprintf("\n%v\n", line)...)
		err = os.Truncate("/etc/hosts", 0)
		if err != nil {
			return err
		}
		err = os.WriteFile("/etc/hosts", b, 0777)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *ETChostmodifier) Reset() error {
	if m.original != nil {
		err := os.Truncate("/etc/hosts", 0)
		if err != nil {
			return err
		}
		err = os.WriteFile("/etc/hosts", m.original, 0644)
		if err != nil {
			return err
		}
	}

	return nil
}
