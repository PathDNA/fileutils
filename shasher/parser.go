package shasher

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
)

// Parse is shorthand for ParseWithToken(nil, r, w).
func Parse(r io.ReadSeeker, w io.Writer) (sig [SignatureSize]byte, bytesWritten int64, err error) {
	return ParseWithToken(nil, r, w)
}

// ParseWithToken parses the r input and verifies it starts with the token then writes the data to w and verifies the signature.
// returns the sha256 signature of the whole message, number of actual data bytes (excluding token and signature) or an err.
func ParseWithToken(token []byte, r io.ReadSeeker, w io.Writer) (sig [SignatureSize]byte, bytesWritten int64, err error) {
	var (
		oVersionAndHash [versionAndHashSize]byte
		sum             = hashAndVersion(version, token)
		h               = sha256.New()
		osig            [SignatureSize]byte
		pos             int64
	)

	if len(token) > 0 {
		if _, err = io.ReadFull(r, oVersionAndHash[:]); err != nil {
			return
		}

		if !bytes.Equal(oVersionAndHash[1:], sum[1:]) {
			err = fmt.Errorf("token mismatch: %x != %x", oVersionAndHash[1:], sum)
			return
		}

		switch oVersionAndHash[0] {
		case version:
		default:
			err = fmt.Errorf("unexpected version: 0x%X", oVersionAndHash[0])
			return
		}

		h.Write(oVersionAndHash[:])
	}

	if pos, err = r.Seek(int64(-SignatureSize), io.SeekEnd); err != nil {
		return
	}

	if _, err = io.ReadFull(r, sig[:]); err != nil {
		err = fmt.Errorf("error reading signature: %v", err)
		return
	}

	if _, err = r.Seek(int64(versionAndHashSize), io.SeekStart); err != nil {
		return
	}

	bytesWritten, err = io.Copy(io.MultiWriter(h, w), io.LimitReader(r, pos-int64(versionAndHashSize)))
	if !bytes.Equal(h.Sum(osig[:0]), sig[:]) {
		err = fmt.Errorf("signature mismatch: %x != %x", osig[:], sig[:])
	}
	return
}
