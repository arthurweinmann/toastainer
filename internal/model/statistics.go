package model

type UserStatistics struct {
	UserID     string  `json:"user_id,omitempty" db:"user_id"`
	Monthyear  string  `json:"month_year,omitempty" db:"month_year"`
	DurationMS int     `json:"duration_ms,omitempty" db:"duration_ms"`
	CPUS       int     `json:"cpu_seconds,omitempty" db:"cpus"`
	Executions int     `json:"runs,omitempty" db:"executions"`
	RAMGBS     float64 `json:"ram_gbs,omitempty" db:"ram_gbs"`
	NetIngress float64 `json:"net_ingress,omitempty" db:"net_ingress"`
	NetEgress  float64 `json:"net_egress,omitempty" db:"net_egress"`
}
