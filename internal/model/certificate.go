package model

type Certificate struct {
	Domain string `json:"domain,omitempty" db:"domain"`
	Cert   []byte `json:"cert,omitempty" db:"cert"`
}
