package diskqueue

import "context"

type qfileSizeReaderInterface interface {
	Read(ctx context.Context, sizeBytes []byte) (err error)
	NextOffset() int64
}

var _ qfileSizeReaderInterface = (*QfileSizeReader)(nil)

// QfileSizeReader for read qfile by size
type QfileSizeReader struct {
	qf         *qfile
	fileOffset int64
	isLatest   bool
}

// if ctx is nil, won't wait for commit
func (r *QfileSizeReader) Read(ctx context.Context, sizeBytes []byte) (err error) {

	_, err = r.qf.mappedFile.ReadRLocked(r.fileOffset, sizeBytes)
	if err != nil {
		if !r.isLatest {
			return
		}
		if ctx == nil {
			// nil means don't wait
			return
		}
		err = r.qf.q.wm.Wait(ctx, r.qf.startOffset+r.fileOffset+int64(len(sizeBytes)))
		if err != nil {
			return
		}
		_, err = r.qf.mappedFile.ReadRLocked(r.fileOffset, sizeBytes)
		if err != nil {
			// 说明换文件了
			return
		}
	}

	r.fileOffset += int64(len(sizeBytes))
	return
}

// NextOffset returns the next offset to read from
func (r *QfileSizeReader) NextOffset() int64 {
	return r.qf.startOffset + r.fileOffset
}
