package goutils

import (
	"crypto/md5"  //#nosec G501 - Not using for crypto
	"crypto/sha1" //#nosec G505 - Not using for crypto
	"crypto/sha256"
	"hash"
	"hash/crc32"
	"io"
)

var (
	castagnoliTable = crc32.MakeTable(crc32.Castagnoli)
)

// MultiHashReader is an io.Reader that hashes the data as its read.
type MultiHashReader struct {
	r          io.Reader
	size       int64
	hashCRC32  hash.Hash32
	hashCRC32C hash.Hash32
	hashMD5    hash.Hash
	hashSHA1   hash.Hash
	hashSHA256 hash.Hash
}

func (r *MultiHashReader) Read(p []byte) (int, error) {
	n, err := r.r.Read(p)
	r.size += int64(n)

	data := p[:n]
	_, _ = r.hashCRC32.Write(data)
	_, _ = r.hashCRC32C.Write(data)
	_, _ = r.hashMD5.Write(data)
	_, _ = r.hashSHA1.Write(data)
	_, _ = r.hashSHA256.Write(data)

	return n, err
}

func (r *MultiHashReader) CRC32() uint32 {
	return r.hashCRC32.Sum32()
}

func (r *MultiHashReader) CRC32C() uint32 {
	return r.hashCRC32C.Sum32()
}

func (r *MultiHashReader) MD5() []byte {
	return r.hashMD5.Sum(nil)
}

func (r *MultiHashReader) SHA1() []byte {
	return r.hashSHA1.Sum(nil)
}

func (r *MultiHashReader) SHA256() []byte {
	return r.hashSHA256.Sum(nil)
}

func (r *MultiHashReader) GetSize() int64 {
	return r.size
}

func NewMultiReader(r io.Reader) *MultiHashReader {
	return &MultiHashReader{
		r:          r,
		size:       0,
		hashCRC32:  crc32.New(crc32.IEEETable),
		hashCRC32C: crc32.New(castagnoliTable),
		//#nosec G401 - this is fine
		hashMD5: md5.New(),
		//#nosec G401 - this is fine
		hashSHA1:   sha1.New(),
		hashSHA256: sha256.New(),
	}
}
