package golite

import (
	"golite/internal/btree"
	"golite/internal/pager"
	"golite/internal/schema"
	"golite/internal/vfs"
)

func Open(filename string) (*DB, error) {
	osVfs := vfs.NewOSVFS()
	file, err := osVfs.Open(filename, 0)
	if err != nil {
		return nil, err
	}
	
	p := pager.New(file, 4096)
	bt := btree.New(p)

	db := &DB{
		backends:   make([]*Backend, 0),
		autoCommit: true,
	}

	mainBackend := &Backend{
		Name:   "main",
		Btree:  bt,
		Schema: &schema.Schema{
			Tables: make(map[string]*schema.Table),
			Indexes: make(map[string]*schema.Index),
		},
	}
	db.backends = append(db.backends, mainBackend)

	return db, nil
}
