package logger

import (
	"fmt"

	"go.uber.org/zap"
)

// EnvType for environment type
type EnvType int

const (
	// Prod for production environment (default)
	Prod EnvType = iota
	// Dev for develop environment
	Dev
)

// Env for chosen environment
var Env EnvType

var (
	logger *zap.Logger
)

// SetInstance for set logger instance
func SetInstance(zl *zap.Logger) {
	logger = zl
}

// Instance for chosen logger
// you can change the actual instance by either:
// 1. change Env
// 2. call SetInstance
func Instance() *zap.Logger {
	if logger != nil {
		return logger
	}
	switch Env {
	case Prod:
		return ProdInstance()
	case Dev:
		return DevInstance()
	default:
		panic(fmt.Sprintf("unknown EnvType:%v", Env))
	}
}
