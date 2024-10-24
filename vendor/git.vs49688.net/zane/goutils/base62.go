package goutils

import (
	"math/big"

	"github.com/google/uuid"
)

func ToBase62(uuid uuid.UUID) string {
	var i big.Int
	i.SetBytes(append([]byte{1}, uuid[:]...))
	return i.Text(62)
}

func FromBase62(s string) (uuid.UUID, bool) {
	var i big.Int
	_, ok := i.SetString(s, 62)
	if !ok {
		return uuid.UUID{}, false
	}

	b := i.Bytes()
	if len(b) != len(uuid.Nil)+1 || b[0] != 1 {
		return uuid.UUID{}, false
	}

	var uuid uuid.UUID
	copy(uuid[:], b[1:])
	return uuid, true
}

func MustFromBase62(s string) uuid.UUID {
	u, ok := FromBase62(s)
	if !ok {
		panic("invalid base62 uuid " + s)
	}

	return u
}
