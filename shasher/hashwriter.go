package shasher

import (
	"crypto/sha256"
	"hash"
	"io"
)

func New(w io.Writer) *HashWriter {
	return &HashWriter{
		w: w,
		h: sha256.New(),
	}
}

func NewWithToken(w io.Writer, token []byte) (*HashWriter, error) {
	hw := &HashWriter{
		w: w,
		h: sha256.New(),
	}
	if _, err := hw.Write(token); err != nil {
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

func (hw *HashWriter) Sign() (sig [32]byte, err error) {
	var n int
	hw.h.Sum(sig[:0])
	if n, err = hw.w.Write(sig[:]); err != nil {
		return
	}
	hw.len += int64(n)
	return sig, err
}

func (hw *HashWriter) Size() int64 { return hw.len }
