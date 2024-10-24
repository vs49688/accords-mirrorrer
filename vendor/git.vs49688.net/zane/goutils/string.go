package goutils

import (
	"crypto/rand"
	"encoding/hex"
)

func RandString(len int) (string, error) {
	b := make([]byte, len)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}

func MustRandString(len int) string {
	s, err := RandString(len)
	if err != nil {
		panic(err)
	}

	return s
}
