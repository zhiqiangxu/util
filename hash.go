package util

import "hash/crc32"

// CRC32 computes the Castagnoli CRC32 of the given data.
func CRC32(data []byte) (sum32 uint32, err error) {
	hash := crc32.New(crc32.MakeTable(crc32.Castagnoli))
	if _, err = hash.Write(data); err != nil {
		return
	}
	sum32 = hash.Sum32()
	return
}
