package bytes

import "unsafe"

const (
	// Align4Mask for 4 bytes boundary
	Align4Mask = 3
	// Align8Mask for 8 bytes boundary
	Align8Mask = 7
)

// AlignedTo4 returns a byte slice aligned to 4 bytes boundary
func AlignedTo4(n uint32) []byte {
	buf := make([]byte, int(n)+Align4Mask)
	buf0Alignment := uint32(uintptr(unsafe.Pointer(&buf[0]))) & uint32(Align4Mask)
	buf = buf[buf0Alignment : buf0Alignment+n]
	return buf
}

// AlignedTo8 returns a byte slice aligned to 8 bytes boundary
func AlignedTo8(n uint32) []byte {
	buf := make([]byte, int(n)+Align8Mask)
	buf0Alignment := uint32(uintptr(unsafe.Pointer(&buf[0]))) & uint32(Align8Mask)
	buf = buf[buf0Alignment : buf0Alignment+n]
	return buf
}
