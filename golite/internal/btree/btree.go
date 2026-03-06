package btree

import (
	"encoding/binary"
	"errors"
	"golite/internal/pager"
	"golite/internal/util"
	"io"
)

// SQLite B-Tree Page Flags
const (
	PTF_INTKEY    = 0x01 // Table B-Tree (integer keys)
	PTF_ZERODATA  = 0x02 // Index B-Tree (no data, keys only)
	PTF_LEAFDATA  = 0x04 // Table B-Tree leaf (contains data)
	PTF_LEAF      = 0x08 // Leaf node
)

// B-Tree creation flags
const (
	BTREE_INTKEY  = 1 
	BTREE_BLOBKEY = 2 
)

type Btree interface {
	BeginTrans(write bool) error
	Commit() error
	Rollback() error

	CreateTable(flags int) (util.Pgno, error)
	DropTable(pgno util.Pgno) error
	Cursor(root util.Pgno, write bool) (Cursor, error)
}

type Cursor interface {
	First() error
	Last() error
	Next() error
	Prev() error
	Seek(key []byte) (bool, error)

	Key() ([]byte, error)
	Data() ([]byte, error)

	Insert(key []byte, data []byte) error
	Delete() error
	Close() error
}

type btree struct {
	p pager.Pager
}

func New(p pager.Pager) Btree {
	return &btree{p: p}
}

func (b *btree) BeginTrans(write bool) error {
	return b.p.Begin(write, false)
}

func (b *btree) Commit() error {
	return b.p.Commit()
}

func (b *btree) Rollback() error {
	return b.p.Rollback()
}

func (b *btree) CreateTable(flags int) (util.Pgno, error) {
	pgno, err := b.p.Allocate()
	if err != nil {
		return 0, err
	}
	
	pg, err := b.p.Get(pgno)
	if err != nil {
		return 0, err
	}
	defer b.p.Release(pg)

	err = b.p.Write(pg)
	if err != nil {
		return 0, err
	}

	data := pg.Data()
	var ptfFlags byte
	if flags&BTREE_INTKEY != 0 {
		ptfFlags = PTF_INTKEY | PTF_LEAFDATA | PTF_LEAF
	} else {
		ptfFlags = PTF_ZERODATA | PTF_LEAF
	}
	
	hdrOff := b.headerOffset(pgno)
	for i := range data {
		data[i] = 0
	}
	data[hdrOff] = ptfFlags
	
	binary.BigEndian.PutUint16(data[hdrOff+5:], uint16(len(data)))
	
	pg.SetDirty()
	return pgno, nil
}

func (b *btree) DropTable(pgno util.Pgno) error {
	return nil
}

func (b *btree) Cursor(root util.Pgno, write bool) (Cursor, error) {
	return &cursor{
		bt:    b,
		root:  root,
		write: write,
		stack: make([]cursorFrame, 0, 20),
	}, nil
}

type cursorFrame struct {
	pg    pager.Page
	index int 
}

type cursor struct {
	bt    *btree
	root  util.Pgno
	write bool
	stack []cursorFrame
}

func (c *cursor) Close() error {
	for _, frame := range c.stack {
		c.bt.p.Release(frame.pg)
	}
	c.stack = nil
	return nil
}

func (c *cursor) getPage(pgno util.Pgno) (pager.Page, error) {
	return c.bt.p.Get(pgno)
}

func (c *cursor) pushPage(pgno util.Pgno) error {
	pg, err := c.getPage(pgno)
	if err != nil {
		return err
	}
	c.stack = append(c.stack, cursorFrame{pg: pg, index: 0})
	return nil
}

func (c *cursor) popPage() {
	if len(c.stack) > 0 {
		frame := c.stack[len(c.stack)-1]
		c.bt.p.Release(frame.pg)
		c.stack = c.stack[:len(c.stack)-1]
	}
}

func (c *cursor) top() *cursorFrame {
	if len(c.stack) == 0 {
		return nil
	}
	return &c.stack[len(c.stack)-1]
}

func getPageHeader(data []byte, pgno util.Pgno) (flags byte, firstFree, nCell, cellOffset uint16, frag byte, rightChild util.Pgno) {
	hdrOff := 0
	if pgno == 1 {
		hdrOff = 100
	}
	flags = data[hdrOff]
	firstFree = binary.BigEndian.Uint16(data[hdrOff+1:])
	nCell = binary.BigEndian.Uint16(data[hdrOff+3:])
	cellOffset = binary.BigEndian.Uint16(data[hdrOff+5:])
	frag = data[hdrOff+7]
	if flags&(PTF_LEAF) == 0 {
		rightChild = util.Pgno(binary.BigEndian.Uint32(data[hdrOff+8:]))
	}
	return
}

func (c *cursor) First() error {
	c.Close()
	if err := c.pushPage(c.root); err != nil {
		return err
	}
	for {
		frame := c.top()
		data := frame.pg.Data()
		flags, _, nCell, _, _, _ := getPageHeader(data, frame.pg.Pgno())
		if flags&PTF_LEAF != 0 {
			if nCell == 0 {
				frame.index = -1
				return io.EOF
			}
			frame.index = 0
			return nil
		}
		if nCell == 0 {
			return errors.New("empty interior node")
		}
		childPgno := binary.BigEndian.Uint32(data[getCellPtr(data, frame.pg.Pgno(), 0):])
		if err := c.pushPage(util.Pgno(childPgno)); err != nil {
			return err
		}
	}
}

func (c *cursor) Last() error {
	c.Close()
	if err := c.pushPage(c.root); err != nil {
		return err
	}
	for {
		frame := c.top()
		data := frame.pg.Data()
		flags, _, nCell, _, _, rightChild := getPageHeader(data, frame.pg.Pgno())
		if flags&PTF_LEAF != 0 {
			if nCell == 0 {
				frame.index = -1
				return io.EOF
			}
			frame.index = int(nCell) - 1
			return nil
		}
		if rightChild == 0 {
			return errors.New("invalid right child in interior node")
		}
		if err := c.pushPage(rightChild); err != nil {
			return err
		}
	}
}

func (c *cursor) Next() error {
	frame := c.top()
	if frame == nil {
		return errors.New("cursor not initialized")
	}
	data := frame.pg.Data()
	flags, _, nCell, _, _, _ := getPageHeader(data, frame.pg.Pgno())

	if flags&PTF_LEAF != 0 {
		frame.index++
		if frame.index < int(nCell) {
			return nil
		}
		for {
			c.popPage()
			frame = c.top()
			if frame == nil {
				return io.EOF
			}
			frame.index++
			data = frame.pg.Data()
			_, _, nCell, _, _, _ = getPageHeader(data, frame.pg.Pgno())
			if frame.index < int(nCell) {
				return nil 
			}
		}
	}
	return errors.New("Next() on interior node not fully implemented")
}

func (c *cursor) Prev() error { return errors.New("not implemented") }

func (c *cursor) Seek(key []byte) (bool, error) {
	return false, errors.New("not implemented")
}

func (c *cursor) Key() ([]byte, error) {
	frame := c.top()
	if frame == nil || frame.index < 0 {
		return nil, errors.New("invalid cursor position")
	}
	info := c.parseCell(frame)
	return info.key, nil
}

func (c *cursor) Data() ([]byte, error) {
	frame := c.top()
	if frame == nil || frame.index < 0 {
		return nil, errors.New("invalid cursor position")
	}
	info := c.parseCell(frame)
	return info.data, nil
}

func (c *cursor) Insert(key []byte, data []byte) error {
	if !c.write {
		return errors.New("cannot insert with read-only cursor")
	}
	
	frame := c.top()
	if frame == nil {
		return errors.New("cursor not positioned")
	}
	
	pg := frame.pg
	pgData := pg.Data()
	_, _, nCell, cellOffset, _, _ := getPageHeader(pgData, pg.Pgno())
	
	cellSize := 2 + len(key) + len(data) 
	usableSpace := int(cellOffset) - (int(getCellPtr(pgData, pg.Pgno(), int(nCell))))
	
	if cellSize > usableSpace {
		if err := c.balance(); err != nil {
			return err
		}
		frame = c.top()
		pg = frame.pg
		pgData = pg.Data()
		_, _, nCell, cellOffset, _, _ = getPageHeader(pgData, pg.Pgno())
	}
	
	err := c.bt.p.Write(pg)
	if err != nil {
		return err
	}
	
	newCellOffset := int(cellOffset) - cellSize
	binary.BigEndian.PutUint16(pgData[c.bt.headerOffset(pg.Pgno())+3:], nCell+1)
	binary.BigEndian.PutUint16(pgData[c.bt.headerOffset(pg.Pgno())+5:], uint16(newCellOffset))
	
	// Write cell pointer
	ptrOff := int(getCellPtr(pgData, pg.Pgno(), int(nCell)))
	binary.BigEndian.PutUint16(pgData[ptrOff:], uint16(newCellOffset))

	// Write cell data: [payloadSize(varint), rowid(varint), payload]
	// key here is the rowid already encoded as varint from VDBE
	n := util.PutVarint(pgData[newCellOffset:], uint64(len(data)))
	copy(pgData[newCellOffset+n:], key)
	copy(pgData[newCellOffset+n+len(key):], data)
	
	pg.SetDirty()
	return nil
}

func (c *cursor) balance() error {
	frame := c.top()
	if frame.pg.Pgno() == c.root {
		return c.balanceDeeper()
	}
	return c.splitPage(frame)
}

func (c *cursor) balanceDeeper() error {
	rootPg := c.top().pg
	newPgno, err := c.bt.p.Allocate()
	if err != nil { return err }
	newPg, err := c.bt.p.Get(newPgno)
	if err != nil { return err }
	defer c.bt.p.Release(newPg)
	
	if err := c.bt.p.Write(newPg); err != nil { return err }
	if err := c.bt.p.Write(rootPg); err != nil { return err }
	
	// Copy root content to new page
	copy(newPg.Data(), rootPg.Data())
	
	// Re-init root as interior node
	data := rootPg.Data()
	hdrOff := c.bt.headerOffset(rootPg.Pgno())
	for i := range data { data[i] = 0 }
	data[hdrOff] = 0x05 // Interior Table B-Tree
	binary.BigEndian.PutUint32(data[hdrOff+8:], uint32(newPgno))
	
	rootPg.SetDirty()
	newPg.SetDirty()
	
	// Update stack
	c.stack = append(c.stack, cursorFrame{pg: newPg, index: 0})
	return nil
}

func (c *cursor) splitPage(frame *cursorFrame) error {
	return errors.New("page split not fully implemented")
}

func (c *cursor) Delete() error { return errors.New("not implemented") }

type cellInfo struct {
	key  []byte
	data []byte
}

func (c *cursor) parseCell(frame *cursorFrame) *cellInfo {
	data := frame.pg.Data()
	ptr := getCellOffset(data, frame.pg.Pgno(), frame.index)
	
	payloadSize, n := util.GetVarint(data[ptr:])
	ptr += n
	rowid, n := util.GetVarint(data[ptr:])
	ptr += n
	
	// Table B-Tree Leaf: [payloadSize, rowid, payload]
	key := make([]byte, 8)
	binary.BigEndian.PutUint64(key, rowid)
	
	return &cellInfo{
		key:  key,
		data: data[ptr : ptr+int(payloadSize)],
	}
}

func getCellOffset(data []byte, pgno util.Pgno, index int) int {
	hdrOff := 0
	if pgno == 1 { hdrOff = 100 }
	ptrOff := hdrOff + 8 + (index * 2)
	if data[hdrOff] == 0x05 || data[hdrOff] == 0x02 { // Interior
		ptrOff = hdrOff + 12 + (index * 2)
	}
	return int(binary.BigEndian.Uint16(data[ptrOff:]))
}

func getCellPtr(data []byte, pgno util.Pgno, index int) int {
	hdrOff := 0
	if pgno == 1 {
		hdrOff = 100
	}
	flags := data[hdrOff]
	ptrSize := 8
	if flags&PTF_LEAF == 0 {
		ptrSize = 12
	}
	return hdrOff + ptrSize + (index * 2)
}

func (b *btree) headerOffset(pgno util.Pgno) int {
	if pgno == 1 {
		return 100
	}
	return 0
}
