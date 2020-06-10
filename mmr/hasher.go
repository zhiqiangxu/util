package mmr

import "crypto/sha256"

// HashType can be replaced by https://github.com/zhiqiangxu/gg
type HashType [32]byte

// Hasher for mmr
type Hasher interface {
	Empty() HashType
	Leaf(data []byte) HashType
	Node(left, right HashType) HashType
}

type hasher32 struct {
	leafPrefix []byte
	nodePrefix []byte
}

// NewHasher creates a Hasher
func NewHasher(leafPrefix, nodePrefix []byte) Hasher {
	return &hasher32{leafPrefix: leafPrefix, nodePrefix: nodePrefix}
}

func (self *hasher32) Empty() HashType {
	return sha256.Sum256(nil)
}

func (self *hasher32) Leaf(leaf []byte) HashType {
	data := append([]byte{}, self.leafPrefix...)
	data = append(data, leaf...)
	return sha256.Sum256(data)
}

func (self *hasher32) Node(left, right HashType) HashType {
	data := append([]byte{}, self.nodePrefix...)
	data = append(data, left[:]...)
	data = append(data, right[:]...)
	return sha256.Sum256(data)
}
