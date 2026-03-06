package pager

import (
	"encoding/binary"
	"golite/internal/util"
	"golite/internal/vfs"
)

const (
	WAL_HDRSIZE       = 32
	WAL_FRAME_HDRSIZE = 24
	WAL_MAGIC         = 0x377f0682 
)

type WalHeader struct {
	Magic     uint32
	Version   uint32
	PageSize  uint32
	CkptSeq   uint32
	Salt      [2]uint32
	Checksum  [2]uint32
}

type WalFrameHeader struct {
	Pgno     util.Pgno
	DbSize   uint32 
	Salt     [2]uint32
	Checksum [2]uint32
}

type Wal struct {
	file     vfs.File
	pageSize uint32
	readOnly bool
	header   WalHeader
	mxFrame  uint32 
}

func NewWal(file vfs.File, pageSize uint32) *Wal {
	return &Wal{
		file:     file,
		pageSize: pageSize,
	}
}

func (w *Wal) checksum(data []byte, s0, s1 uint32, native bool) (uint32, uint32) {
	for i := 0; i < len(data); i += 8 {
		var x0, x1 uint32
		if native {
			x0 = binary.BigEndian.Uint32(data[i : i+4])
			x1 = binary.BigEndian.Uint32(data[i+4 : i+8])
		} else {
			x0 = binary.LittleEndian.Uint32(data[i : i+4])
			x1 = binary.LittleEndian.Uint32(data[i+4 : i+8])
		}
		s0 += x0 + s1
		s1 += x1 + s0
	}
	return s0, s1
}

func (w *Wal) WriteFrames(pages []Page, dbSize util.Pgno) error {
	return nil 
}

func (w *Wal) Checkpoint(dbFile vfs.File) error {
	return nil 
}

func (w *Wal) FindFrame(pgno util.Pgno) (uint32, error) {
	return 0, nil
}

func (w *Wal) ReadFrame(iFrame uint32, data []byte) error {
	off := int64(WAL_HDRSIZE) + int64(iFrame-1)*int64(WAL_FRAME_HDRSIZE+w.pageSize) + WAL_FRAME_HDRSIZE
	_, err := w.file.ReadAt(data, off)
	return err
}
