package record

import (
	"bytes"
	"testing"
)

func TestEncodeDecodeRecord(t *testing.T) {
	vals := []Value{
		NewNullValue(),
		NewIntValue(1),
		NewIntValue(100),
		NewIntValue(100000),
		NewStringValue("hello"),
		NewBlobValue([]byte{1, 2, 3}),
	}

	data := EncodeRecord(vals)

	for i, want := range vals {
		got, err := DecodeColumn(data, i)
		if err != nil {
			t.Fatalf("DecodeColumn(%d) failed: %v", i, err)
		}
		
		if got.Type() != want.Type() {
			t.Errorf("Column %d: type = %d, want %d", i, got.Type(), want.Type())
		}
		
		switch want.Type() {
		case 1:
			if got.Int64() != want.Int64() {
				t.Errorf("Column %d: %d, want %d", i, got.Int64(), want.Int64())
			}
		case 3:
			if got.String() != want.String() {
				t.Errorf("Column %d: %q, want %q", i, got.String(), want.String())
			}
		case 4:
			if !bytes.Equal(got.Blob(), want.Blob()) {
				t.Errorf("Column %d: %x, want %x", i, got.Blob(), want.Blob())
			}
		}
	}
}
