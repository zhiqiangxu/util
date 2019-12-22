package diskqueue

// Conf for diskqueue
type Conf struct {
	Directory          string
	WriteBatch         int
	WriteMmap          bool
	MaxMsgSize         int
	MaxPutting         int
	ByteArenaChunkSize int
	EnableWriteBuffer  bool
	// only valid when EnableWriteBuffer is true
	// unit: second
	CommitInterval int
}
