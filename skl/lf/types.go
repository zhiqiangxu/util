package lf

type list interface {
	Contains(k []byte) bool
	Get(k []byte) (v []byte, exists bool)
	Insert(k, v []byte) (isNew bool, err error)
	Delete(k []byte) bool
}
