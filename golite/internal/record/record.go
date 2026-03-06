package record

import (
	"encoding/binary"
	"fmt"
	"golite/internal/util"
	"math"
)

type Value interface {
	Type() byte // 1: Int, 2: Float, 3: Text, 4: Blob, 5: Null
	Int64() int64
	Float64() float64
	String() string
	Blob() []byte
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
func (v *stringValue) Float64() float64 { return 0.0 }
func (v *stringValue) String() string { return v.val }
func (v *stringValue) Blob() []byte { return []byte(v.val) }
func (v *stringValue) Type() byte { return 3 }

type blobValue struct{ val []byte }
func (v *blobValue) Int64() int64 { return 0 }
func (v *blobValue) Float64() float64 { return 0.0 }
func (v *blobValue) String() string { return string(v.val) }
func (v *blobValue) Blob() []byte { return v.val }
func (v *blobValue) Type() byte { return 4 }

type nullValue struct{}
func (v *nullValue) Int64() int64 { return 0 }
func (v *nullValue) Float64() float64 { return 0.0 }
func (v *nullValue) String() string { return "NULL" }
func (v *nullValue) Blob() []byte { return nil }
func (v *nullValue) Type() byte { return 5 }

func NewIntValue(v int64) Value { return &intValue{val: v} }
func NewFloatValue(v float64) Value { return &floatValue{val: v} }
func NewStringValue(v string) Value { return &stringValue{val: v} }
func NewBlobValue(v []byte) Value { return &blobValue{val: v} }
func NewNullValue() Value { return &nullValue{} }

func serialType(val Value) uint64 {
	if val == nil || val.Type() == 5 {
		return 0
	}
	switch val.Type() {
	case 1:
		v := val.Int64()
		if v >= 0 && v <= 1 {
			return uint64(8 + v) 
		} else if v >= math.MinInt8 && v <= math.MaxInt8 {
			return 1
		} else if v >= math.MinInt16 && v <= math.MaxInt16 {
			return 2
		} else if v >= -8388608 && v <= 8388607 {
			return 3 // 24-bit
		} else if v >= math.MinInt32 && v <= math.MaxInt32 {
			return 4
		} else if v >= -140737488355328 && v <= 140737488355327 {
			return 5 // 48-bit
		}
		return 6
	case 2:
		return 7
	case 4:
		return uint64(12 + len(val.Blob())*2)
	case 3:
		return uint64(13 + len(val.String())*2)
	}
	return 0
}

func EncodeRecord(values []Value) []byte {
	serialTypes := make([]uint64, len(values))
	headerSize := uint64(0)
	payloadSize := 0

	for i, v := range values {
		st := serialType(v)
		serialTypes[i] = st
		headerSize += uint64(util.VarintLen(st))
		
		switch st {
		case 1: payloadSize += 1
		case 2: payloadSize += 2
		case 3: payloadSize += 3
		case 4: payloadSize += 4
		case 5: payloadSize += 6
		case 6: payloadSize += 8
		case 7: payloadSize += 8
		default:
			if st >= 12 && st%2 == 0 {
				payloadSize += int((st - 12) / 2)
			} else if st >= 13 && st%2 == 1 {
				payloadSize += int((st - 13) / 2)
			}
		}
	}

	headerSizeVarintLen := util.VarintLen(headerSize + uint64(util.VarintLen(headerSize)))
	headerSize += uint64(headerSizeVarintLen) 
	
	record := make([]byte, int(headerSize)+payloadSize)
	n := util.PutVarint(record, headerSize)
	
	for _, st := range serialTypes {
		n += util.PutVarint(record[n:], st)
	}

	for i, v := range values {
		st := serialTypes[i]
		switch st {
		case 0, 8, 9:
		case 1:
			record[n] = byte(v.Int64())
			n += 1
		case 2:
			binary.BigEndian.PutUint16(record[n:], uint16(v.Int64()))
			n += 2
		case 3:
			val := int32(v.Int64())
			record[n] = byte(val >> 16)
			record[n+1] = byte(val >> 8)
			record[n+2] = byte(val)
			n += 3
		case 4:
			binary.BigEndian.PutUint32(record[n:], uint32(v.Int64()))
			n += 4
		case 5:
			val := v.Int64()
			record[n] = byte(val >> 40)
			record[n+1] = byte(val >> 32)
			record[n+2] = byte(val >> 24)
			record[n+3] = byte(val >> 16)
			record[n+4] = byte(val >> 8)
			record[n+5] = byte(val)
			n += 6
		case 6:
			binary.BigEndian.PutUint64(record[n:], uint64(v.Int64()))
			n += 8
		case 7:
			binary.BigEndian.PutUint64(record[n:], math.Float64bits(v.Float64()))
			n += 8
		default:
			if st >= 12 && st%2 == 0 {
				b := v.Blob()
				copy(record[n:], b)
				n += len(b)
			} else if st >= 13 && st%2 == 1 {
				s := v.String()
				copy(record[n:], s)
				n += len(s)
			}
		}
	}
	
	return record
}

func DecodeColumn(data []byte, colIndex int) (Value, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty record")
	}

	hdrSize, n := util.GetVarint(data)
	if n == 0 || int(hdrSize) > len(data) {
		return nil, fmt.Errorf("invalid header size")
	}

	payloadOffset := int(hdrSize)
	currentHdrOffset := n

	for i := 0; i <= colIndex; i++ {
		if currentHdrOffset >= int(hdrSize) {
			return &nullValue{}, nil
		}
		
		st, n := util.GetVarint(data[currentHdrOffset:])
		currentHdrOffset += n
		
		var colLen int
		switch st {
		case 0, 8, 9: colLen = 0
		case 1: colLen = 1
		case 2: colLen = 2
		case 3: colLen = 3
		case 4: colLen = 4
		case 5: colLen = 6
		case 6, 7: colLen = 8
		default:
			if st >= 12 && st%2 == 0 {
				colLen = int((st - 12) / 2)
			} else if st >= 13 && st%2 == 1 {
				colLen = int((st - 13) / 2)
			}
		}
		
		if i == colIndex {
			if payloadOffset+colLen > len(data) {
				return nil, fmt.Errorf("corrupt record, payload out of bounds")
			}
			
			switch st {
			case 0: return &nullValue{}, nil
			case 1: return &intValue{val: int64(int8(data[payloadOffset]))}, nil
			case 2: return &intValue{val: int64(int16(binary.BigEndian.Uint16(data[payloadOffset:])))}, nil
			case 3: 
				v := uint32(data[payloadOffset])<<16 | uint32(data[payloadOffset+1])<<8 | uint32(data[payloadOffset+2])
				if v&0x800000 != 0 { v |= 0xff000000 }
				return &intValue{val: int64(int32(v))}, nil
			case 4: return &intValue{val: int64(int32(binary.BigEndian.Uint32(data[payloadOffset:])))}, nil
			case 5:
				v := uint64(data[payloadOffset])<<40 | uint64(data[payloadOffset+1])<<32 | uint64(data[payloadOffset+2])<<24 | uint64(data[payloadOffset+3])<<16 | uint64(data[payloadOffset+4])<<8 | uint64(data[payloadOffset+5])
				if v&0x800000000000 != 0 { v |= 0xffff000000000000 }
				return &intValue{val: int64(v)}, nil
			case 6: return &intValue{val: int64(binary.BigEndian.Uint64(data[payloadOffset:]))}, nil
			case 7: return &floatValue{val: math.Float64frombits(binary.BigEndian.Uint64(data[payloadOffset:]))}, nil
			case 8: return &intValue{val: 0}, nil
			case 9: return &intValue{val: 1}, nil
			default:
				if st >= 12 && st%2 == 0 {
					return &blobValue{val: append([]byte(nil), data[payloadOffset:payloadOffset+colLen]...)}, nil
				} else if st >= 13 && st%2 == 1 {
					return &stringValue{val: string(data[payloadOffset:payloadOffset+colLen])}, nil
				}
			}
		}
		payloadOffset += colLen
	}
	return nil, fmt.Errorf("column index out of range")
}
