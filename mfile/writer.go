package mfile

import (
	"io"
	"os"
)

// TODO: should writer implement Reader as well?

var (
	_ io.Seeker      = (*Writer)(nil)
	_ io.WriterAt    = (*Writer)(nil)
	_ io.WriteCloser = (*Writer)(nil)
)

// Writer returns `append ? f.WriterAt(f.Size()) : f.WriterAt(0)`.
func (f *File) Writer(append bool) *Writer {
	if append {
		return f.WriterAt(f.size)
	}
	return f.WriterAt(0)
}

// WriterAt acquires a write-lock, seeks to the given offset and returns a writer.
func (f *File) WriterAt(off int64) *Writer {
	f.mux.Lock()

	f.f.Seek(off, io.SeekStart)

	return &Writer{
		f:   f,
		off: off,
	}
}

// Writer implements `io.Writer`, `io.WriterAt`, `io.Seeker`` and `io.Closer`.
type Writer struct {
	f   *File
	off int64
}

// Write implements `io.Writer`.
func (w *Writer) Write(b []byte) (n int, err error) {
	if w.f == nil {
		return 0, os.ErrClosed
	}
	n, err = w.f.f.Write(b)
	w.off += int64(n)
	return
}

// WriteAt implements `io.WriterAt`.
func (w *Writer) WriteAt(b []byte, off int64) (n int, err error) {
	if w.f == nil {
		return 0, os.ErrClosed
	}
	return w.f.f.WriteAt(b, off)
}

// Seek implements `io.Seeker`.
func (w *Writer) Seek(offset int64, whence int) (n int64, err error) {
	if w.f == nil {
		return 0, os.ErrClosed
	}
	n, err = w.f.f.Seek(offset, whence)
	w.off = n
	return
}

// Close releases the parent's write-lock.
func (w *Writer) Close() error {
	if w.f == nil {
		return os.ErrClosed
	}

	err := w.f.closeWriter()
	w.f = nil
	return err
}
