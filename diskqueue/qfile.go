package diskqueue

import (
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/zhiqiangxu/util/logger"
	"github.com/zhiqiangxu/util/mapped"
	"go.uber.org/zap"
)

type qfileInterface interface {
	Shrink() error
	writeBuffers(buffs *net.Buffers) (int64, error)
	WrotePosition() int64
	Write([]byte) (int, error)
	Sync() error
	Close() error
}

type qfile struct {
	qm          *queueMeta
	idx         int
	startOffset int64
	mappedFile  *mapped.File
}

const (
	qfSubDir  = "qf"
	qfileSize = 1024 * 1024 * 1024
)

var _ qfileInterface = (*qfile)(nil)

func qfilePath(startOffset int64, conf *Conf) string {
	return filepath.Join(conf.Directory, qfSubDir, fmt.Sprintf("%20d", startOffset))
}

func openQfile(qm *queueMeta, idx int) (qf *qfile, err error) {
	fm := qm.FileMeta(idx)

	qf = &qfile{qm: qm, idx: idx, startOffset: fm.StartOffset}
	qf.mappedFile, err = mapped.OpenFile(qfilePath(fm.StartOffset, qm.conf), int64(fm.EndOffset-fm.StartOffset), os.O_RDWR, false, nil)
	return
}

func createQfile(qm *queueMeta, idx int, startOffset int64) (qf *qfile, err error) {
	qf = &qfile{qm: qm, idx: idx, startOffset: startOffset}
	qf.mappedFile, err = mapped.CreateFile(qfilePath(startOffset, qm.conf), qfileSize, false, nil)
	if err != nil {
		return
	}

	if qm.NumFiles() != idx {
		logger.Instance().Fatal("createQfile idx != NumFiles", zap.Int("NumFiles", qm.NumFiles()), zap.Int("idx", idx))
	}

	nowNano := NowNano()
	qm.AddFile(FileMeta{StartOffset: startOffset, EndOffset: startOffset, StartTime: nowNano, EndTime: nowNano})
	return
}

func (qf *qfile) writeBuffers(buffs *net.Buffers) (n int64, err error) {
	n, err = buffs.WriteTo(qf.mappedFile)

	return
}

func (qf *qfile) WrotePosition() int64 {
	return qf.mappedFile.WrotePositionNonAtomic()
}

func (qf *qfile) Write(data []byte) (n int, err error) {
	logger.Instance().Fatal("qfile.Write not supported")
	return
}

func (qf *qfile) Shrink() error {
	return qf.mappedFile.Shrink()
}

func (qf *qfile) Sync() error {
	return nil
}

func (qf *qfile) Close() error {
	return nil
}
