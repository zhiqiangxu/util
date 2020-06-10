package logger

import (
	"fmt"
	"sync"
	"sync/atomic"
	"unsafe"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	prodLoggerPtr unsafe.Pointer
	prodMU        sync.Mutex
)

// ProdInstance returns the instance for production environment
func ProdInstance() *zap.Logger {
	prodLogger := atomic.LoadPointer(&prodLoggerPtr)
	if prodLogger != nil {
		return (*zap.Logger)(prodLogger)
	}

	prodMU.Lock()
	defer prodMU.Unlock()
	prodLogger = atomic.LoadPointer(&prodLoggerPtr)
	if prodLogger != nil {
		return (*zap.Logger)(prodLogger)
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeDuration = zapcore.StringDurationEncoder

	zconf := zap.Config{
		DisableCaller:     true,
		DisableStacktrace: true,
		Level:             zap.NewAtomicLevelAt(zapcore.InfoLevel),
		Development:       false,
		Encoding:          "json",
		EncoderConfig:     encoderConfig,
		OutputPaths:       []string{"stdout"},
		ErrorOutputPaths:  []string{"stderr"},
	}

	prodLoggerType, err := New(zconf)
	if err != nil {
		panic(fmt.Sprintf("ProdInstance New:%v", err))
	}

	atomic.StorePointer(&prodLoggerPtr, unsafe.Pointer(prodLoggerType))

	return prodLoggerType
}
