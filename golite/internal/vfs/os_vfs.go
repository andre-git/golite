package vfs

import (
	"os"
)

type osVFS struct{}

func NewOSVFS() VFS {
	return &osVFS{}
}

func (v *osVFS) Open(name string, flags int) (File, error) {
	f, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	return &osFile{f: f}, nil
}

func (v *osVFS) Delete(name string, syncDir bool) error {
	return os.Remove(name)
}

func (v *osVFS) Access(name string, flags int) (bool, error) {
	_, err := os.Stat(name)
	return err == nil, nil
}

type osFile struct {
	f *os.File
}

func (f *osFile) ReadAt(p []byte, off int64) (int, error) {
	return f.f.ReadAt(p, off)
}

func (f *osFile) WriteAt(p []byte, off int64) (int, error) {
	return f.f.WriteAt(p, off)
}

func (f *osFile) Truncate(size int64) error {
	return f.f.Truncate(size)
}

func (f *osFile) Sync(flags int) error {
	return f.f.Sync()
}

func (f *osFile) FileSize() (int64, error) {
	info, err := f.f.Stat()
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

func (f *osFile) Lock(lockType int) error {
	return nil // Basic implementation, no locking yet
}

func (f *osFile) Unlock(lockType int) error {
	return nil
}

func (f *osFile) Close() error {
	return f.f.Close()
}
