package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.SugaredLogger

// Init 初始化日志
func Init(level, format string) {
	var cfg zap.Config

	if format == "json" {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	switch level {
	case "debug":
		cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		cfg.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		cfg.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	logger, err := cfg.Build(zap.AddCallerSkip(1))
	if err != nil {
		panic(err)
	}

	log = logger.Sugar()
}

// Debug 调试日志
func Debug(args ...interface{}) {
	log.Debug(args...)
}

// Debugf 格式化调试日志
func Debugf(template string, args ...interface{}) {
	log.Debugf(template, args...)
}

// Info 信息日志
func Info(args ...interface{}) {
	log.Info(args...)
}

// Infof 格式化信息日志
func Infof(template string, args ...interface{}) {
	log.Infof(template, args...)
}

// Warn 警告日志
func Warn(args ...interface{}) {
	log.Warn(args...)
}

// Warnf 格式化警告日志
func Warnf(template string, args ...interface{}) {
	log.Warnf(template, args...)
}

// Error 错误日志
func Error(args ...interface{}) {
	log.Error(args...)
}

// Errorf 格式化错误日志
func Errorf(template string, args ...interface{}) {
	log.Errorf(template, args...)
}

// Fatalf 致命错误日志
func Fatalf(template string, args ...interface{}) {
	log.Fatalf(template, args...)
}
