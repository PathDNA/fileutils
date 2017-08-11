package shasher

import (
	"crypto/sha256"
	"hash"
	"io"
)

const (
	SignatureSize      = sha256.Size
	versionAndHashSize = 1 + SignatureSize
	version            = 1
)

func New(w io.Writer) *HashWriter {
	return &HashWriter{
		w: w,
		h: sha256.New(),
	}
}

func NewWithToken(w io.Writer, token []byte) (*HashWriter, error) {
	var (
		hw = &HashWriter{
			w: w,
			h: sha256.New(),
		}
		versionAndHash = hashAndVersion(version, token)
	)

	if _, err := hw.Write(versionAndHash); err != nil {
		return nil, err
	}

	return hw, nil
}

type HashWriter struct {
	w   io.Writer
	h   hash.Hash
	len int64
}

func (hw *HashWriter) Write(p []byte) (int, error) {
	hw.h.Write(p)
	n, err := hw.w.Write(p)
	hw.len += int64(n)
	return n, err
}

func (hw *HashWriter) HashOnly(p []byte) (int, error) {
	return hw.h.Write(p)
}

func (hw *HashWriter) Sign() (sig [SignatureSize]byte, err error) {
	var n int
	hw.h.Sum(sig[:0])
	if n, err = hw.w.Write(sig[:]); err != nil {
		return
	}
	hw.len += int64(n)
	return
}

func (hw *HashWriter) Size() int64 { return hw.len }

func hashAndVersion(version uint8, b []byte) []byte {
	var (
		h   = sha256.New()
		sum [versionAndHashSize]byte
	)
	sum[0] = version
	h.Write(b)
	h.Sum(sum[1:1])
	return sum[:]
}
