package diskqueue

type queueInterface interface {
	Put([]byte) (int64, error)
	Read(offset int64, stores [][]byte) (results [][]byte, err error)
	Close() error
	Depth() int64
	Delete() error
}

var _ queueInterface = (*Queue)(nil)

// Queue for diskqueue
type Queue struct {
	meta queueMeta
}

// Put data to queue
func (q *Queue) Put(data []byte) (offset int64, err error) {
	return
}

// ReadFrom for read from offset
func (q *Queue) Read(offset int64, stores [][]byte) (results [][]byte, err error) {
	return
}

// Close the queue
func (q *Queue) Close() error {
	return nil
}

// Depth of the queue
func (q *Queue) Depth() int64 {
	return 0
}

// Delete the queue
func (q *Queue) Delete() error {
	return nil
}
