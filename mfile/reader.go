package mfile

import (
	"fmt"
	"io"
	"os"
)

var (
	_ io.Seeker     = (*Reader)(nil)
	_ io.ReaderAt   = (*Reader)(nil)
	_ io.ReadCloser = (*Reader)(nil)
)

// Reader returns ReaderAt(0).
func (f *File) Reader() *Reader { return f.ReaderAt(0) }

// ReaderAt returns a new reader at the specified offset.
// Close must be called or it will leak.
func (f *File) ReaderAt(off int64) *Reader {
	f.mux.RLock()
	f.wg.Add(1)
	return &Reader{
		f:    f,
		off:  off,
		size: f.getSize(),
	}
}

// Reader implements `io.Reader`, `io.ReaderAt`, `io.Seeker` and `io.Closer`.
type Reader struct {
	f    *File
	off  int64
	size int64
}

// Read implements `io.Read`.
func (r *Reader) Read(b []byte) (n int, err error) {
	n, err = r.ReadAt(b, r.off)
	r.off += int64(n)
	return
}

// ReadAt implements `io.ReaderAt`.
func (r *Reader) ReadAt(b []byte, off int64) (n int, err error) {
	if r.f == nil {
		return 0, os.ErrClosed
	}

	if off >= r.size {
		return 0, io.EOF
	}

	if diff := r.size - r.off; len(b) > int(diff) {
		b = b[:diff]
		if n, err = r.f.f.ReadAt(b, off); err == nil {
			err = io.EOF
		}
	} else {
		n, err = r.f.f.ReadAt(b, off)
	}
	return
}

// Seek implements `io.Seeker`.
func (r *Reader) Seek(offset int64, whence int) (int64, error) {
	if r.f == nil {
		return 0, os.ErrClosed
	}

	switch whence {
	case io.SeekStart:
		r.off = offset
	case io.SeekCurrent:
		r.off += offset
	case io.SeekEnd:
		r.off = r.size - offset
	}

	if r.off < -1 || r.off > r.f.size {
		return 0, fmt.Errorf("%d is invalid (range: 0, %d)", r.off, r.f.size)
	}

	return r.off, nil
}

// Close releases the parent's read-lock.
func (r *Reader) Close() error {
	if r.f == nil {
		return os.ErrClosed
	}
	r.f.wg.Done()
	r.f.mux.RUnlock()
	r.f = nil
	return nil
}
