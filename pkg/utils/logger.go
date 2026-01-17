package utils

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func InitLogger(path string, debug bool) (*zap.Logger, error) {
	// Create log directory if not exists
	if path != "" {
		if err := os.MkdirAll(path, 0755); err != nil {
			return nil, err
		}
	}

	// Encoder config
	encoderConfig := zap.NewProductionEncoderConfig()
	if debug {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
	}
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.CallerKey = "caller"                      // Track file:line
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder // EncodeCaller type short

	// Choose encoder format
	encoder := zapcore.NewJSONEncoder(encoderConfig)
	if debug {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// Set log level
	logLevel := zap.InfoLevel
	if debug {
		logLevel = zap.DebugLevel
	}

	// File sink with log rotation
	logFile := path + time.Now().Format("20060102") + ".log"
	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    10, // MB
		MaxBackups: 7,
		MaxAge:     28, // days
		Compress:   true,
	})

	// Console sink (stdout)
	consoleWriter := zapcore.AddSync(os.Stdout)

	// Combine multiple sinks
	core := zapcore.NewTee(
		zapcore.NewCore(encoder, fileWriter, logLevel),
		zapcore.NewCore(encoder, consoleWriter, logLevel),
	)

	// Create logger with caller info
	logger := zap.New(core, zap.AddCaller())
	return logger, nil
}
