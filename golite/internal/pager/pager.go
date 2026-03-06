package pager

import (
	"fmt"
	"golite/internal/util"
	"golite/internal/vfs"
	"io"
	"sync"
)

type Pager interface {
    Get(pgno util.Pgno) (Page, error)
    Write(pg Page) error
    Release(pg Page)
    
    Begin(writable bool, exclusive bool) error
    Commit() error
    Rollback() error

    SetPageSize(size uint32) error
    PageSize() uint32
}

type Page interface {
    Pgno() util.Pgno
    Data() []byte
    SetDirty()
}

type pager struct {
	mu         sync.Mutex
	file       vfs.File
	pageSize   uint32
	dbSize     util.Pgno
	cache      map[util.Pgno]*page
	bufferPool *sync.Pool
	writable   bool
	exclusive  bool
}

type page struct {
	pgno  util.Pgno
	data  []byte
	dirty bool
	pager *pager
	refs  int
}

func (p *page) Pgno() util.Pgno { return p.pgno }
func (p *page) Data() []byte     { return p.data }
func (p *page) SetDirty()        { p.dirty = true }

func New(file vfs.File, pageSize uint32) Pager {
	if pageSize == 0 {
		pageSize = 4096
	}
	p := &pager{
		file:     file,
		pageSize: pageSize,
		cache:    make(map[util.Pgno]*page),
	}
	p.bufferPool = &sync.Pool{
		New: func() interface{} {
			return make([]byte, p.pageSize)
		},
	}
	// Determine dbSize from file size
	fs, _ := file.FileSize()
	p.dbSize = util.Pgno(fs / int64(pageSize))
	return p
}

func (p *pager) Get(pgno util.Pgno) (Page, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if pg, ok := p.cache[pgno]; ok {
		pg.refs++
		return pg, nil
	}

	data := p.bufferPool.Get().([]byte)
	
	// Read from file
	off := int64(pgno-1) * int64(p.pageSize)
	n, err := p.file.ReadAt(data, off)
	if err != nil && err != io.EOF {
		p.bufferPool.Put(data)
		return nil, err
	}
	
	// Zero out the rest of the buffer if we hit EOF or partial read
	if n < len(data) {
		for i := n; i < len(data); i++ {
			data[i] = 0
		}
	}

	pg := &page{
		pgno:  pgno,
		data:  data,
		pager: p,
		refs:  1,
	}
	p.cache[pgno] = pg
	return pg, nil
}

func (p *pager) Write(pg Page) error {
	if !p.writable {
		return fmt.Errorf("pager is not in a write transaction")
	}
	pg.SetDirty()
	return nil
}

func (p *pager) Release(pg Page) {
	p.mu.Lock()
	defer p.mu.Unlock()

	pgObj := pg.(*page)
	pgObj.refs--
}

func (p *pager) Begin(writable bool, exclusive bool) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.writable = writable
	p.exclusive = exclusive
	return nil
}

func (p *pager) Commit() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.writable {
		return fmt.Errorf("not in a write transaction")
	}

	for pgno, pg := range p.cache {
		if pg.dirty {
			off := int64(pgno-1) * int64(p.pageSize)
			_, err := p.file.WriteAt(pg.data, off)
			if err != nil {
				return err
			}
			pg.dirty = false
		}
	}

	err := p.file.Sync(0)
	if err != nil {
		return err
	}

	p.writable = false
	return nil
}

func (p *pager) Rollback() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for pgno, pg := range p.cache {
		if pg.dirty {
			delete(p.cache, pgno)
			p.bufferPool.Put(pg.data)
		}
	}
	p.writable = false
	return nil
}

func (p *pager) SetPageSize(size uint32) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.cache) > 0 {
		return fmt.Errorf("cannot change page size while pages are in cache")
	}
	p.pageSize = size
	p.bufferPool = &sync.Pool{
		New: func() interface{} {
			return make([]byte, size)
		},
	}
	return nil
}

func (p *pager) PageSize() uint32 {
	return p.pageSize
}
