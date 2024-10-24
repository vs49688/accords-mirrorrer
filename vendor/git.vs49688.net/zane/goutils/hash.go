package goutils

import (
	"encoding/hex"
	"hash"
)

func Hash(hp func() hash.Hash, data []byte) []byte {
	h := hp()
	_, _ = h.Write(data)
	return h.Sum(nil)
}

func MustDecodeHex(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}

	return b
}
