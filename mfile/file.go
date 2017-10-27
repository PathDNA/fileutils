package mfile

import (
	"io"
	"os"
	"sync"
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
		f:      f,
		closed: make(chan struct{}),
		size:   sz,
	}, nil
}

// File is an *os.File wrapper that allows multiple readers or a single writer.
type File struct {
	f   *os.File
	mux sync.RWMutex

	closed chan struct{}
	size   int64
	locked bool
}

func (f *File) Reader() *Reader { return f.ReaderAt(0) }

func (f *File) ReaderAt(off int64) *Reader {
	f.mux.RLock()
	return &Reader{
		f:   f,
		off: off,
	}
}

func (f *File) Writer(append bool) *Writer {
	if append {
		return f.WriterAt(f.size)
	}
	return f.WriterAt(0)
}

func (f *File) WriterAt(off int64) *Writer {
	f.mux.Lock()

	f.f.Seek(off, io.SeekStart)

	return &Writer{
		f:   f,
		off: off,
	}
}

func (f *File) Truncate(sz int64) (err error) {
	f.mux.Lock()
	if err = f.f.Truncate(sz); err == nil {
		f.size = sz
	}
	f.mux.Unlock()
	return
}

func (f *File) Size() int64 {
	f.mux.RLock()
	sz := f.size
	f.mux.RUnlock()
	return sz
}

func (f *File) With(fn func(*os.File) error) error {
	f.mux.Lock()
	defer f.mux.Unlock()
	return fn(f.f)
}

func (f *File) closeWriter() (err error) {
	defer f.mux.Unlock()
	if err = f.f.Sync(); err != nil {
		return
	}

	f.size, err = getSize(f.f)
	return
}

func (f *File) ForceClose() error {
	if f.locked {
		f.unlock()
	}
	return f.f.Close()
}

// Close signals all the underlying readers and writers and waits for them to finish then closes the actual file.
func (f *File) Close() error {
	f.mux.Lock()
	if f.locked {
		f.unlock()
	}
	err := f.f.Close()
	f.mux.Unlock()

	return err
}

func getSize(f *os.File) (int64, error) {
	st, err := f.Stat()

	if err != nil {
		return 0, err
	}

	return st.Size(), nil
}
