package vdbe

import (
	"fmt"
	"golite/internal/btree"
	"golite/internal/record"
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
	OP_Add          = 14
	OP_Subtract     = 15
	OP_Multiply     = 16
	OP_Divide       = 17
	OP_Remainder    = 18
	OP_Eq           = 19
	OP_Ne           = 20
	OP_Lt           = 21
	OP_Le           = 22
	OP_Gt           = 23
	OP_Ge           = 24
	OP_NewRowid     = 25
	OP_MakeRecord   = 26
	OP_Insert       = 27
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

type blobValue struct{ val []byte }
func (v *blobValue) Int64() int64 { return 0 }
func (v *blobValue) String() string { return string(v.val) }
func (v *blobValue) Type() byte { return 4 }

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
		case OP_CreateTable:
			pgno, err := v.bt.CreateTable(btree.BTREE_INTKEY)
			if err != nil {
				return 0, err
			}
			v.registers[pOp.P2] = &intValue{val: int64(pgno)}
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
		case OP_Add:
			v1 := v.registers[pOp.P1].Int64()
			v2 := v.registers[pOp.P2].Int64()
			v.registers[pOp.P3] = &intValue{val: v2 + v1}
		case OP_Subtract:
			v1 := v.registers[pOp.P1].Int64()
			v2 := v.registers[pOp.P2].Int64()
			v.registers[pOp.P3] = &intValue{val: v2 - v1}
		case OP_Multiply:
			v1 := v.registers[pOp.P1].Int64()
			v2 := v.registers[pOp.P2].Int64()
			v.registers[pOp.P3] = &intValue{val: v2 * v1}
		case OP_Divide:
			v1 := v.registers[pOp.P1].Int64()
			v2 := v.registers[pOp.P2].Int64()
			if v1 != 0 {
				v.registers[pOp.P3] = &intValue{val: v2 / v1}
			} else {
				v.registers[pOp.P3] = nil 
			}
		case OP_Remainder:
			v1 := v.registers[pOp.P1].Int64()
			v2 := v.registers[pOp.P2].Int64()
			if v1 != 0 {
				v.registers[pOp.P3] = &intValue{val: v2 % v1}
			} else {
				v.registers[pOp.P3] = nil
			}
		case OP_Eq:
			v1 := v.registers[pOp.P1].Int64()
			v3 := v.registers[pOp.P3].Int64()
			if v3 == v1 {
				v.pc = pOp.P2
			}
		case OP_Ne:
			v1 := v.registers[pOp.P1].Int64()
			v3 := v.registers[pOp.P3].Int64()
			if v3 != v1 {
				v.pc = pOp.P2
			}
		case OP_Lt:
			v1 := v.registers[pOp.P1].Int64()
			v3 := v.registers[pOp.P3].Int64()
			if v3 < v1 {
				v.pc = pOp.P2
			}
		case OP_Le:
			v1 := v.registers[pOp.P1].Int64()
			v3 := v.registers[pOp.P3].Int64()
			if v3 <= v1 {
				v.pc = pOp.P2
			}
		case OP_Gt:
			v1 := v.registers[pOp.P1].Int64()
			v3 := v.registers[pOp.P3].Int64()
			if v3 > v1 {
				v.pc = pOp.P2
			}
		case OP_Ge:
			v1 := v.registers[pOp.P1].Int64()
			v3 := v.registers[pOp.P3].Int64()
			if v3 >= v1 {
				v.pc = pOp.P2
			}
		case OP_NewRowid:
			cur := v.cursors[pOp.P1]
			if cur == nil {
				return 0, fmt.Errorf("invalid cursor %d", pOp.P1)
			}
			var newId int64 = 1
			err := cur.Last()
			if err == nil {
				key, err := cur.Key()
				if err == nil && len(key) > 0 {
					lastId, _ := util.GetVarint(key)
					newId = int64(lastId) + 1
				}
			}
			v.registers[pOp.P2] = &intValue{val: newId}

		case OP_MakeRecord:
			firstReg := pOp.P1
			nField := pOp.P2
			destReg := pOp.P3
			values := make([]record.Value, nField)
			for i := 0; i < nField; i++ {
				val := v.registers[firstReg+i]
				if val == nil {
					values[i] = record.NewNullValue()
					continue
				}
				switch val.Type() {
				case 1:
					values[i] = record.NewIntValue(val.Int64())
				case 3:
					values[i] = record.NewStringValue(val.String())
				case 4:
					values[i] = record.NewBlobValue([]byte(val.String()))
				default:
					values[i] = record.NewNullValue()
				}
			}
			rec := record.EncodeRecord(values)
			v.registers[destReg] = &blobValue{val: rec}

		case OP_Insert:
			cur := v.cursors[pOp.P1]
			if cur == nil {
				return 0, fmt.Errorf("invalid cursor %d", pOp.P1)
			}
			recVal := v.registers[pOp.P2]
			rowidVal := v.registers[pOp.P3]
			if recVal == nil || rowidVal == nil {
				return 0, fmt.Errorf("OP_Insert: record or rowid register is null")
			}
			rowid := uint64(rowidVal.Int64())
			key := make([]byte, 10)
			n := util.PutVarint(key, rowid)
			
			var data []byte
			if bv, ok := recVal.(*blobValue); ok {
				data = bv.val
			} else {
				data = []byte(recVal.String())
			}
			if err := cur.Insert(key[:n], data); err != nil {
				return 0, err
			}
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
