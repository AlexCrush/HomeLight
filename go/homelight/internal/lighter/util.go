package lighter

import (
	"fmt"
)

func bytesToString(b []byte) string {
	res := ""
	for _, v := range b {
		res += fmt.Sprintf("%X ", v)
	}
	return res
}

func shrinkData(data []byte) ([]byte, error) {
	if len(data)%2 != 0 {
		return nil, fmt.Errorf("non-even length of input data")
	}
	shortData := make([]byte, len(data)/2)
	mask := byte(0xF0)
	for i := 0; i < len(data); i += 2 {
		bH := data[i]
		if bH&0xF0 != mask {
			return nil, fmt.Errorf("invalid byte %X at pos %d", bH, i)
		}
		mask = 0
		bL := data[i+1]
		if bL&0xF0 != 0x00 {
			return nil, fmt.Errorf("invalid byte %X at pos %d", bL, i)
		}
		b := ((bH & 0xF) << 4) | bL
		shortData[i/2] = b
	}
	return shortData, nil
}

func widenData(data []byte) []byte {
	res := make([]byte, len(data)*2)
	mask := byte(0xF0)
	for i, v := range data {
		res[i*2] = ((v & 0xF0) >> 4) | mask
		res[i*2+1] = v & 0x0F
		mask = 0
	}
	return res
}

func lampPosAndMask(lampId int) (int, byte) {
	pos := lampId / 8
	mask := byte(1 << (lampId % 8))
	return pos, mask
}
