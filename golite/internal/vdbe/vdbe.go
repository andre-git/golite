package vdbe

import (
	"bytes"
	"fmt"
	"golite/internal/btree"
	"golite/internal/record"
	"golite/internal/util"
	"sort"
)

const (
	OP_Noop           = 0
	OP_Halt           = 1
	OP_Integer        = 2
	OP_String         = 3
	OP_Goto           = 4
	OP_ResultRow      = 5
	OP_Column         = 6
	OP_Transaction    = 7
	OP_CreateTable    = 8
	OP_OpenRead       = 9
	OP_OpenWrite      = 10
	OP_Rewind         = 11
	OP_Next           = 12
	OP_SeekGE         = 13
	OP_Add            = 14
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
	OP_OpenEphemeral  = 28
	OP_SorterOpen     = 29
	OP_SorterInsert   = 30
	OP_SorterSort     = 31
	OP_SorterData     = 32
	OP_SorterNext     = 33
	OP_If             = 34
	OP_IfNot          = 35
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
	Float64() float64
	String() string
	Blob() []byte
	Type() byte
}

type intValue struct{ val int64 }
func (v *intValue) Int64() int64 { return v.val }
func (v *intValue) Float64() float64 { return float64(v.val) }
func (v *intValue) String() string { return fmt.Sprintf("%d", v.val) }
func (v *intValue) Blob() []byte { return nil }
func (v *intValue) Type() byte { return 1 }

type floatValue struct{ val float64 }
func (v *floatValue) Int64() int64 { return int64(v.val) }
func (v *floatValue) Float64() float64 { return v.val }
func (v *floatValue) String() string { return fmt.Sprintf("%g", v.val) }
func (v *floatValue) Blob() []byte { return nil }
func (v *floatValue) Type() byte { return 2 }

type stringValue struct{ val string }
func (v *stringValue) Int64() int64 { return 0 }
func (v *stringValue) Float64() float64 { return 0 }
func (v *stringValue) String() string { return v.val }
func (v *stringValue) Blob() []byte { return []byte(v.val) }
func (v *stringValue) Type() byte { return 3 }

type blobValue struct{ val []byte }
func (v *blobValue) Int64() int64 { return 0 }
func (v *blobValue) Float64() float64 { return 0 }
func (v *blobValue) String() string { return string(v.val) }
func (v *blobValue) Blob() []byte { return v.val }
func (v *blobValue) Type() byte { return 4 }

type Sorter struct {
	records [][]byte
	current int
}

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
	sorters   map[int]*Sorter
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
		sorters:   make(map[int]*Sorter),
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
			val, err := record.DecodeColumn(data, pOp.P2)
			if err != nil {
				return 0, err
			}
			switch val.Type() {
			case 1: v.registers[pOp.P3] = &intValue{val: val.Int64()}
			case 2: v.registers[pOp.P3] = &floatValue{val: val.Float64()}
			case 3: v.registers[pOp.P3] = &stringValue{val: val.String()}
			case 4: v.registers[pOp.P3] = &blobValue{val: val.Blob()}
			case 5: v.registers[pOp.P3] = nil 
			default: v.registers[pOp.P3] = &stringValue{val: val.String()}
			}
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
			v1 := v.registers[pOp.P1]
			v3 := v.registers[pOp.P3]
			if (v1 == nil && v3 == nil) || (v1 != nil && v3 != nil && v1.String() == v3.String()) {
				v.pc = pOp.P2
			}
		case OP_Ne:
			v1 := v.registers[pOp.P1]
			v3 := v.registers[pOp.P3]
			var s1, s3 string
			if v1 != nil { s1 = v1.String() }
			if v3 != nil { s3 = v3.String() }
			if (v1 == nil && v3 != nil) || (v1 != nil && v3 == nil) || (v1 != nil && v3 != nil && s1 != s3) {
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
				case 1: values[i] = record.NewIntValue(val.Int64())
				case 2: values[i] = record.NewFloatValue(val.Float64())
				case 3: values[i] = record.NewStringValue(val.String())
				case 4: values[i] = record.NewBlobValue(val.Blob())
				default: values[i] = record.NewNullValue()
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
		case OP_OpenEphemeral:
			pgno, err := v.bt.CreateTable(btree.BTREE_BLOBKEY)
			if err != nil { return 0, err }
			cur, err := v.bt.Cursor(util.Pgno(pgno), true)
			if err != nil { return 0, err }
			v.closeCursor(pOp.P1)
			v.cursors[pOp.P1] = cur
		case OP_SorterOpen:
			v.sorters[pOp.P1] = &Sorter{records: [][]byte{}, current: -1}
		case OP_SorterInsert:
			s := v.sorters[pOp.P1]
			val := v.registers[pOp.P2]
			if bv, ok := val.(*blobValue); ok {
				s.records = append(s.records, bv.val)
			}
		case OP_SorterSort:
			s := v.sorters[pOp.P1]
			if len(s.records) == 0 {
				v.pc = pOp.P2
				continue
			}
			sort.SliceStable(s.records, func(i, j int) bool {
				return bytes.Compare(s.records[i], s.records[j]) < 0
			})
			s.current = 0
		case OP_SorterData:
			s := v.sorters[pOp.P1]
			v.registers[pOp.P2] = &blobValue{val: s.records[s.current]}
		case OP_SorterNext:
			s := v.sorters[pOp.P1]
			s.current++
			if s.current < len(s.records) {
				v.pc = pOp.P2
			}
		case OP_If:
			val := v.registers[pOp.P1]
			if val != nil && val.Int64() != 0 {
				v.pc = pOp.P2
			}
		case OP_IfNot:
			val := v.registers[pOp.P1]
			if val == nil || val.Int64() == 0 {
				v.pc = pOp.P2
			}
		default:
			return 0, fmt.Errorf("unsupported opcode: %d at PC %d", pOp.Op, v.pc-1)
		}
	}
	v.halted = true
	return 101, nil 
}

func (v *vdbeState) closeCursor(i int) {
	if i < len(v.cursors) && v.cursors[i] != nil {
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
	v.sorters = make(map[int]*Sorter)
	return nil
}

func (v *vdbeState) Finalize() error {
	v.Reset()
	v.ops = nil
	v.registers = nil
	v.halted = true
	return nil
}
