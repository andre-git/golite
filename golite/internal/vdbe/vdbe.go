package vdbe

import (
	"fmt"
	"golite/internal/btree"
	"golite/internal/util"
)

const (
	OP_Noop         = 0
	OP_Halt         = 1
	OP_Integer      = 2
	OP_String       = 3
	OP_Goto         = 4
	OP_ResultRow    = 5
	OP_Column       = 6
	OP_Transaction  = 7
	OP_CreateTable  = 8
	OP_OpenRead     = 9
	OP_OpenWrite    = 10
	OP_Rewind       = 11
	OP_Next         = 12
	OP_SeekGE       = 13
)

type Opcode struct {
	Op uint8
	P1 int
	P2 int
	P3 int
	P4 any
	P5 uint16
}

type Value interface {
	Int64() int64
	String() string
	Type() byte
}

type intValue struct{ val int64 }
func (v *intValue) Int64() int64 { return v.val }
func (v *intValue) String() string { return fmt.Sprintf("%d", v.val) }
func (v *intValue) Type() byte { return 1 }

type stringValue struct{ val string }
func (v *stringValue) Int64() int64 { return 0 }
func (v *stringValue) String() string { return v.val }
func (v *stringValue) Type() byte { return 3 }

type Vdbe interface {
	Step() (int, error)
	Reset() error
	Finalize() error
}

type vdbeState struct {
	bt        btree.Btree
	ops       []Opcode
	pc        int
	registers []Value
	cursors   []btree.Cursor
	halted    bool
	rc        int
}

func NewVdbe(bt btree.Btree, ops []Opcode, nMem int, nCursor int) Vdbe {
	return &vdbeState{
		bt:        bt,
		ops:       ops,
		pc:        0,
		registers: make([]Value, nMem),
		cursors:   make([]btree.Cursor, nCursor),
	}
}

func (v *vdbeState) Step() (int, error) {
	if v.halted {
		return 0, fmt.Errorf("VDBE halted")
	}

	for v.pc >= 0 && v.pc < len(v.ops) {
		pOp := &v.ops[v.pc]
		v.pc++

		switch pOp.Op {
		case OP_Noop:
			continue
		case OP_Halt:
			v.halted = true
			return pOp.P1, nil
		case OP_Integer:
			v.registers[pOp.P2] = &intValue{val: int64(pOp.P1)}
		case OP_String:
			v.registers[pOp.P2] = &stringValue{val: pOp.P4.(string)}
		case OP_Goto:
			v.pc = pOp.P2
		case OP_ResultRow:
			return 100, nil // SQLITE_ROW
		case OP_Transaction:
			if err := v.bt.BeginTrans(pOp.P2 != 0); err != nil {
				return 0, err
			}
		case OP_OpenRead:
			cur, err := v.bt.Cursor(util.Pgno(pOp.P2), false)
			if err != nil {
				return 0, err
			}
			v.closeCursor(pOp.P1)
			v.cursors[pOp.P1] = cur
		case OP_OpenWrite:
			cur, err := v.bt.Cursor(util.Pgno(pOp.P2), true)
			if err != nil {
				return 0, err
			}
			v.closeCursor(pOp.P1)
			v.cursors[pOp.P1] = cur
		case OP_Rewind:
			cur := v.cursors[pOp.P1]
			if cur == nil {
				return 0, fmt.Errorf("invalid cursor %d", pOp.P1)
			}
			err := cur.First()
			if err != nil {
				v.pc = pOp.P2 
			}
		case OP_Next:
			cur := v.cursors[pOp.P1]
			if cur == nil {
				return 0, fmt.Errorf("invalid cursor %d", pOp.P1)
			}
			err := cur.Next()
			if err == nil {
				v.pc = pOp.P2
			}
		case OP_Column:
			cur := v.cursors[pOp.P1]
			if cur == nil {
				return 0, fmt.Errorf("invalid cursor %d", pOp.P1)
			}
			data, err := cur.Data()
			if err != nil {
				return 0, err
			}
			v.registers[pOp.P3] = &stringValue{val: string(data)}
		default:
			return 0, fmt.Errorf("unsupported opcode: %d at PC %d", pOp.Op, v.pc-1)
		}
	}
	v.halted = true
	return 101, nil // SQLITE_DONE
}

func (v *vdbeState) closeCursor(i int) {
	if v.cursors[i] != nil {
		v.cursors[i].Close()
		v.cursors[i] = nil
	}
}

func (v *vdbeState) Reset() error {
	v.pc = 0
	v.halted = false
	for i := range v.registers {
		v.registers[i] = nil
	}
	for i := range v.cursors {
		v.closeCursor(i)
	}
	return nil
}

func (v *vdbeState) Finalize() error {
	v.Reset()
	v.ops = nil
	v.registers = nil
	v.halted = true
	return nil
}
