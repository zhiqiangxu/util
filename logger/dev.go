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
	devLoggerPtr unsafe.Pointer
	devMU        sync.Mutex
)

// DevInstance returns the instance for develop environment
func DevInstance() *zap.Logger {
	devLogger := atomic.LoadPointer(&devLoggerPtr)
	if devLogger != nil {
		return (*zap.Logger)(devLogger)
	}

	devMU.Lock()
	defer devMU.Unlock()
	devLogger = atomic.LoadPointer(&devLoggerPtr)
	if devLogger != nil {
		return (*zap.Logger)(devLogger)
	}

	encoderConfig := zap.NewDevelopmentEncoderConfig()

	zconf := zap.Config{
		DisableCaller:     true,
		DisableStacktrace: true,
		Level:             zap.NewAtomicLevelAt(zapcore.DebugLevel),
		Development:       true,
		Encoding:          "json",
		EncoderConfig:     encoderConfig,
		OutputPaths:       []string{"stdout"},
		ErrorOutputPaths:  []string{"stderr"},
	}

	devLoggerType, err := New(zconf)
	if err != nil {
		panic(fmt.Sprintf("DevInstance New:%v", err))
	}

	atomic.StorePointer(&devLoggerPtr, unsafe.Pointer(devLoggerType))

	return devLoggerType
}
