package diskqueue

// Conf for diskqueue
type Conf struct {
	Directory         string
	WriteBatch        int
	MaxMsgSize        int
	EnableWriteBuffer bool
}
