package config

import "encoding/json"

func DefaultConfig() map[string]interface{} {
	def := &loadedConfig{
		IsAPI:    true,
		IsRunner: true,
	}

	b, _ := json.Marshal(def)
	m := map[string]interface{}{}
	json.Unmarshal(b, &m)

	return m
}
