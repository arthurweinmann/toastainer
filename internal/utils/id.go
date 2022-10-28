package utils

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/rs/xid"
)

func UniqueSecureID120() (string, error) {
	b := make([]byte, 100)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b)[:100] + xid.New().String(), nil
}

func UniqueSecureID60() (string, error) {
	b := make([]byte, 40)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b)[:40] + xid.New().String(), nil
}

func UniqueSecureID36() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b)[:16] + xid.New().String(), nil
}

func SecureToken32() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b)[:32], nil
}
