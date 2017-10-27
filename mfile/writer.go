package mfile

import (
	"io"
	"os"
)

// TODO: should writer implement Reader as well?

var (
	_ io.WriteSeeker = (*Writer)(nil)
	_ io.WriterAt    = (*Writer)(nil)
)

// Writer implements io.Writer, io.WriterAt, io.Seeker and io.Closer
type Writer struct {
	f   *File
	off int64
}

func (w *Writer) Write(b []byte) (n int, err error) {
	if w.f == nil {
		return 0, os.ErrClosed
	}
	n, err = w.f.f.Write(b)
	w.off += int64(n)
	return
}

func (w *Writer) WriteAt(b []byte, off int64) (n int, err error) {
	if w.f == nil {
		return 0, os.ErrClosed
	}
	return w.f.f.WriteAt(b, off)
}

func (w *Writer) Seek(offset int64, whence int) (n int64, err error) {
	if w.f == nil {
		return 0, os.ErrClosed
	}
	n, err = w.f.f.Seek(offset, whence)
	w.off = n
	return
}

func (w *Writer) Close() error {
	if w.f == nil {
		return os.ErrClosed
	}

	err := w.f.closeWriter()
	w.f = nil
	return err
}
