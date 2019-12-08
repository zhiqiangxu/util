package diskqueue

// Conf for diskqueue
type Conf struct {
	Directory          string
	WriteBatch         int
	MaxMsgSize         int
	MaxPutting         int
	ByteArenaChunkSize int
	EnableWriteBuffer  bool
}
