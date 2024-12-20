package utils

import (
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"strings"
	"time"
)

const (
	Debug = 1
	Info  = 2
	Warn  = -1
	Error = -2
)

func Logger(name string) logr.Logger {
	config := zap.Config{
		Encoding:          "json",
		Level:             zap.NewAtomicLevelAt(GetLogLevel()),
		OutputPaths:       []string{"stdout"},
		ErrorOutputPaths:  []string{"stderr"},
		DisableStacktrace: false,
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:    "message",
			NameKey:       "logger",
			LevelKey:      "level",
			EncodeLevel:   zapcore.CapitalLevelEncoder,
			TimeKey:       "time",
			StacktraceKey: "stacktrace",
			EncodeTime: func(time time.Time, encoder zapcore.PrimitiveArrayEncoder) {
				encoder.AppendString(time.Format("2006-01-02T15:04:05.999"))
			},
		},
	}

	logger, err := config.Build()
	if err != nil {
		panic("Cannot initialize logger. Reason: " + err.Error())
	}

	return zapr.NewLogger(logger).WithName(name)
}

func GetLogLevel() zapcore.Level {
	switch strings.ToLower(os.Getenv("LOG_LEVEL")) {
	case "debug":
		return zap.DebugLevel
	case "error":
		return zap.ErrorLevel
	case "warn":
		return zap.WarnLevel
	case "info":
	default:
		return zapcore.InfoLevel
	}

	return zapcore.InfoLevel
}
