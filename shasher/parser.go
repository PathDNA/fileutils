package shasher

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
)

// Parse is shorthand for ParseWithToken(nil, r, w).
func Parse(r io.ReadSeeker, w io.Writer) (sig [sha256.Size]byte, bytesWritten int64, err error) {
	return ParseWithToken(nil, r, w)
}

// ParseWithToken parses the r input and verifies it starts with the token then writes the data to w and verifies the signature.
// returns the sha256 signature of the whole message, number of actual data bytes (excluding token and signature) or an err.
func ParseWithToken(token []byte, r io.ReadSeeker, w io.Writer) (sig [sha256.Size]byte, bytesWritten int64, err error) {
	var (
		otoken = make([]byte, len(token))
		h      = sha256.New()
		osig   [sha256.Size]byte
		pos    int64
	)

	if len(token) > 0 {
		if _, err = io.ReadFull(r, otoken); err != nil {
			return
		}

		if !bytes.Equal(token, otoken) {
			err = fmt.Errorf("token mismatch: %s != %s", token, otoken)
			return
		}

		h.Write(token)
	}

	if pos, err = r.Seek(int64(-sha256.Size), io.SeekEnd); err != nil {
		return
	}

	if _, err = io.ReadFull(r, sig[:]); err != nil {
		err = fmt.Errorf("error reading signature: %v", err)
		return
	}

	if _, err = r.Seek(int64(len(token)), io.SeekStart); err != nil {
		return
	}

	bytesWritten, err = io.Copy(io.MultiWriter(h, w), io.LimitReader(r, pos-int64(len(token))))
	if !bytes.Equal(h.Sum(osig[:0]), sig[:]) {
		err = fmt.Errorf("signature mismatch: %x != %x", osig[:], sig[:])
	}
	return
}
