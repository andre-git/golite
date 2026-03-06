package util

func PutVarint(p []byte, v uint64) int {
	if v <= 0x7f {
		p[0] = byte(v)
		return 1
	}
	if v <= 0x3fff {
		p[0] = byte((v >> 7) | 0x80)
		p[1] = byte(v & 0x7f)
		return 2
	}
	
	// Fast path for 9 bytes
	if v > 0x00ffffffffffffff {
		p[0] = byte((v>>56)&0x7f | 0x80)
		p[1] = byte((v>>49)&0x7f | 0x80)
		p[2] = byte((v>>42)&0x7f | 0x80)
		p[3] = byte((v>>35)&0x7f | 0x80)
		p[4] = byte((v>>28)&0x7f | 0x80)
		p[5] = byte((v>>21)&0x7f | 0x80)
		p[6] = byte((v>>14)&0x7f | 0x80)
		p[7] = byte((v>>7)&0x7f | 0x80)
		p[8] = byte(v & 0xff) // Note: 9th byte uses all 8 bits
		return 9
	}
	
	var buf [10]byte
	n := 0
	for {
		buf[n] = byte((v & 0x7f) | 0x80)
		n++
		v >>= 7
		if v == 0 {
			break
		}
	}
	buf[0] &= 0x7f // Clear continuation bit of the last byte
	
	for i := 0; i < n; i++ {
		p[i] = buf[n-1-i]
	}
	
	return n
}

func GetVarint(p []byte) (uint64, int) {
	var v uint64
	for i := 0; i < 9 && i < len(p); i++ {
		b := p[i]
		if i < 8 {
			v = (v << 7) | uint64(b&0x7f)
			if b&0x80 == 0 {
				return v, i + 1
			}
		} else {
			v = (v << 8) | uint64(b)
			return v, 9
		}
	}
	return v, len(p)
}

func VarintLen(v uint64) int {
	if v <= 0x7f { return 1 }
	if v <= 0x3fff { return 2 }
	if v <= 0x1fffff { return 3 }
	if v <= 0xfffffff { return 4 }
	if v <= 0x7ffffffff { return 5 }
	if v <= 0x3ffffffffff { return 6 }
	if v <= 0x1ffffffffffff { return 7 }
	if v <= 0xffffffffffffff { return 8 }
	return 9
}
