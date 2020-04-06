package mmr

import "crypto/sha256"

type hasher32 struct {
	nodePrefix []byte
}

// NewHasher creates a Hasher
func NewHasher(nodePrefix []byte) Hasher {
	return &hasher32{nodePrefix: nodePrefix}
}

func (self *hasher32) Empty() HashType {
	return sha256.Sum256(nil)
}

func (self *hasher32) Node(left, right HashType) HashType {
	data := append([]byte{}, self.nodePrefix...)
	data = append(data, left[:]...)
	data = append(data, right[:]...)
	return sha256.Sum256(data)
}
