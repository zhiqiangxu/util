package bytes

import "unsafe"

const (
	align4Mask = 3
	align8Mask = 7
)

// AlignedTo4 returns a byte slice aligned to 4 byte boundary
func AlignedTo4(n uint32) []byte {
	buf := make([]byte, int(n)+align4Mask)
	buf0Alignment := uint32(uintptr(unsafe.Pointer(&buf[0]))) & uint32(align4Mask)
	buf = buf[buf0Alignment : buf0Alignment+n]
	return buf
}

// AlignedTo8 returns a byte slice aligned to 8 byte boundary
func AlignedTo8(n uint32) []byte {
	buf := make([]byte, int(n)+align8Mask)
	buf0Alignment := uint32(uintptr(unsafe.Pointer(&buf[0]))) & uint32(align8Mask)
	buf = buf[buf0Alignment : buf0Alignment+n]
	return buf
}
