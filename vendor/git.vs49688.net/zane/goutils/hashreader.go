package goutils

import (
	"hash"
	"io"
)

// HashReader is an io.Reader that hashes the data as its read.
type HashReader struct {
	r    io.Reader
	size int64
	hash hash.Hash
}

func (r *HashReader) Read(p []byte) (int, error) {
	n, err := r.r.Read(p)
	r.size += int64(n)

	data := p[:n]
	_, _ = r.hash.Write(data)

	return n, err
}

func (r *HashReader) Hash() []byte {
	return r.hash.Sum(nil)
}

func (r *HashReader) GetSize() int64 {
	return r.size
}

func NewHashReader(r io.Reader, h func() hash.Hash) *HashReader {
	return &HashReader{
		r:    r,
		size: 0,
		hash: h(),
	}
}
