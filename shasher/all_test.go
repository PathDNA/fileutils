package shasher

import (
	"bytes"
	"testing"
)

func TestParser(t *testing.T) {
	var (
		tok = []byte("omg this is a token")
		in  bytes.Buffer
		out bytes.Buffer
	)

	hw, _ := NewWithToken(&in, tok)
	hw.Write([]byte("actual data"))
	sig, err := hw.Sign()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("signature: %x", sig)

	psig, _, err := ParseWithToken(tok, bytes.NewReader(in.Bytes()), &out)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(psig[:], sig[:]) {
		t.Fatalf("sig mismatch: %x != %x", sig, psig)
	}

	if v := out.String(); v != "actual data" {
		t.Fatalf("unexpected data: %s", v)
	}
}
