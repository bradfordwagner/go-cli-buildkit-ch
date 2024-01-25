package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.SugaredLogger

func init() {
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	c := zap.Config{
		EncoderConfig:    encoderCfg,
		Encoding:         "console",
		ErrorOutputPaths: []string{"stderr"},
		Level:            zap.NewAtomicLevelAt(zap.InfoLevel),
		OutputPaths:      []string{"stderr"},
	}
	l, _ := c.Build()
	logger = l.Sugar()
}

func Log() *zap.SugaredLogger {
	return logger
}
