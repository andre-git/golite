package util

import (
	"bytes"
	"testing"
)

func TestVarint(t *testing.T) {
	tests := []struct {
		val  uint64
		want []byte
	}{
		{0, []byte{0x00}},
		{127, []byte{0x7f}},
		{128, []byte{0x81, 0x00}},
		{16383, []byte{0xff, 0x7f}},
		{16384, []byte{0x81, 0x80, 0x00}},
		{0xffffffffffffffff, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
	}

	for _, tt := range tests {
		buf := make([]byte, 10)
		n := PutVarint(buf, tt.val)
		if n != len(tt.want) {
			t.Errorf("PutVarint(%d) returned length %d, want %d", tt.val, n, len(tt.want))
		}
		if !bytes.Equal(buf[:n], tt.want) {
			t.Errorf("PutVarint(%d) = %x, want %x", tt.val, buf[:n], tt.want)
		}

		v, n2 := GetVarint(buf[:n])
		if v != tt.val {
			t.Errorf("GetVarint(%x) = %d, want %d", buf[:n], v, tt.val)
		}
		if n2 != n {
			t.Errorf("GetVarint consumed %d bytes, want %d", n2, n)
		}
	}
}
