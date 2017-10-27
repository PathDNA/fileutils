package mfile

// TODO:
// * explore file locking
// * use mmap, but that would make appending/trunc a lot harder.

import (
	"io"
	"os"
	"sync"
)

var (
	_ io.ReaderFrom = (*File)(nil)
	_ io.ReaderAt   = (*File)(nil)
	_ io.WriterTo   = (*File)(nil)
	_ io.WriterAt   = (*File)(nil)
	_ io.Closer     = (*File)(nil)
)

// New opens fp with os.O_CREATE|os.O_RDWR and the given permissions.
func New(fp string, perm os.FileMode) (*File, error) {
	f, err := os.OpenFile(fp, os.O_CREATE|os.O_RDWR, perm)

	if err != nil {
		return nil, err
	}

	sz, err := getSize(f)

	if err != nil {
		return nil, err
	}

	return &File{
		f:    f,
		size: sz,
	}, nil
}

// File is an `*os.File` wrapper that allows multiple readers or one writer on a single file descriptor.
type File struct {
	f   *os.File
	mux sync.RWMutex

	size int64

	SyncAfterWriterClose bool // if set to true, calling `Writer.Close()`, will call `*os.File.Sync()`.
}

// ReadFrom implements `io.ReaderFrom`.
func (f *File) ReadFrom(rd io.Reader) (n int64, err error) {
	wr := f.WriterAt(0)
	n, err = io.Copy(wr, rd)
	wr.Close()
	return
}

// WriteTo implements `io.WriterTo`.
func (f *File) WriteTo(w io.Writer) (n int64, err error) {
	r := f.ReaderAt(0)
	n, err = io.Copy(w, r)
	r.Close()
	return
}

// ReadAt implements `io.ReaderAt`.
func (f *File) ReadAt(b []byte, off int64) (n int, err error) {
	r := f.ReaderAt(off)
	n, err = r.Read(b)
	return
}

// WriteAt implements `io.WriterAt`.
func (f *File) WriteAt(b []byte, off int64) (n int, err error) {
	wr := f.WriterAt(off)
	n, err = wr.Write(b)
	wr.Close()
	return
}

// Append wraps `f.Writer(true)` for convience.
func (f *File) Append(b []byte) (n int, err error) {
	w := f.Writer(true)
	n, err = w.Write(b)
	w.Close()
	return
}

// Truncate truncates the underlying `*os.File` to the specific size.
func (f *File) Truncate(sz int64) (err error) {
	f.mux.Lock()
	if err = f.f.Truncate(sz); err == nil {
		f.size = sz
	}
	f.mux.Unlock()
	return
}

// Size returns the current file size.
// the size is cached after each writer is closed, so it doesn't call Stat().
func (f *File) Size() int64 {
	f.mux.RLock()
	sz := f.size
	f.mux.RUnlock()
	return sz
}

// Stat calls the underlying `*os.File.Stat()`.
func (f *File) Stat() (fi os.FileInfo, err error) {
	f.mux.RLock()
	fi, err = f.f.Stat()
	f.mux.RUnlock()
	return
}

// With acquires a write lock and calls fn with the underlying `*os.File` and returns any errors it returns.
func (f *File) With(fn func(*os.File) error) error {
	f.mux.Lock()
	defer f.mux.Unlock()
	return fn(f.f)
}

func (f *File) closeWriter() (err error) {
	defer f.mux.Unlock()

	if f.SyncAfterWriterClose {
		if err = f.f.Sync(); err != nil {
			return
		}
	}

	f.size, err = getSize(f.f)
	return
}

// ForceClose will close the underlying `*os.File` without waiting for any active readers/writer.
func (f *File) ForceClose() error {
	return f.f.Close()
}

// Close waits for all the active readers/writer to finish before closing the underlying `*os.File`.
func (f *File) Close() error {
	f.mux.Lock()
	err := f.f.Close()
	f.mux.Unlock()

	return err
}

func getSize(f *os.File) (int64, error) {
	st, err := f.Stat()

	if err != nil {
		return 0, err
	}

	if !st.Mode().IsRegular() {
		return 0, os.ErrInvalid
	}

	return st.Size(), nil
}
