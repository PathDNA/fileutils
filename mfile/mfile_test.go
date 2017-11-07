package mfile_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"sync"
	"testing"

	"github.com/PathDNA/fileutils/mfile"
	. "github.com/PathDNA/testutils"
)

var dummyData = bytes.Repeat([]byte("0123456789"), 2)

func Test(t *testing.T) {
	f, err := ioutil.TempFile("", "mfile-")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()

	var n int
	if n, err = f.Write(dummyData); FailIf(t, err) || n != len(dummyData) {
		return
	}

	mf, err := mfile.FromFile(f)
	if FailIf(t, err) {
		return
	}

	t.Run("Reader Limit", func(t *testing.T) {
		r := mf.Reader()
		a := mf.Appender()
		a.Write(dummyData)
		a.Close()
		b, err := ioutil.ReadAll(r)
		r.Close()
		if err != nil || !bytes.Equal(b, dummyData) {
			t.Errorf("unexpected read (%s | %s): %v", b, dummyData, err)
		}

		r = mf.Reader()
		b, err = ioutil.ReadAll(r)
		r.Close()
		if err != nil || len(b) != len(dummyData)*2 {
			t.Errorf("invalid file state :( (%s | %s): %v", b, dummyData, err)
		}
	})

	t.Run("Concurrent Read Append", func(t *testing.T) {
		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(2)
			i := i
			go func() {
				defer wg.Done()
				wc := mf.Appender()
				defer wc.Close()
				buf := bytes.Repeat([]byte{byte(i) + '0'}, 10)
				if n, err := wc.Write(buf); err != nil || n != len(buf) {
					t.Errorf("%d (%d/%d): %v", i, n, len(buf), err)
					return
				}
			}()

			go func() {
				defer wg.Done()
				buf := make([]byte, len(dummyData)-10)
				r := mf.ReaderAt(10)
				defer r.Close()
				n, err := r.Read(buf)
				if err != nil || n != len(buf) {
					t.Errorf("%d (%d/%d): %v", i, n, len(buf), err)
					return
				}
				if !bytes.Equal(buf[:n], dummyData[10:]) {
					t.Errorf("%s mismatch", buf)
				}
			}()
		}

		wg.Wait()
	})

	var buf bytes.Buffer
	mf.WriteTo(&buf)
	if buf.Len() != (len(dummyData)*2)+(10*10) {
		t.Errorf("unexpected data: %s", buf.Bytes())
	}

	mf.Close() // make sure there are no deadlocks
}
