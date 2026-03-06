package vfs

type VFS interface {
	Open(name string, flags int) (File, error)
	Delete(name string, syncDir bool) error
	Access(name string, flags int) (bool, error)
}

type File interface {
	ReadAt(p []byte, off int64) (int, error)
	WriteAt(p []byte, off int64) (int, error)
	Truncate(size int64) error
	Sync(flags int) error
	FileSize() (int64, error)
	Lock(lockType int) error
	Unlock(lockType int) error
	Close() error
}
