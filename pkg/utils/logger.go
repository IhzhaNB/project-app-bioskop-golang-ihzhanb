package utils

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func InitLogger(path string, debug bool) (*zap.Logger, error) {
	// Buat folder log jika belum ada
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
	encoderConfig.CallerKey = "caller"                      // TAMBAH INI untuk tau file:line
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder // TAMBAH INI

	// set format log
	encoder := zapcore.NewJSONEncoder(encoderConfig)
	if debug {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// Level log
	logLevel := zap.InfoLevel
	if debug {
		logLevel = zap.DebugLevel
	}

	// File sink dengan rotasi log
	logFile := path + "cinema-booking.log"
	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    10, // MB
		MaxBackups: 7,
		MaxAge:     28, // days
		Compress:   true,
	})

	// Stdout sink
	consoleWriter := zapcore.AddSync(os.Stdout)

	// Gabungkan ke dalam satu core
	core := zapcore.NewTee(
		zapcore.NewCore(encoder, fileWriter, logLevel),
		zapcore.NewCore(encoder, consoleWriter, logLevel),
	)

	// Buat logger DENGAN CALLER (ini yang penting!)
	logger := zap.New(core, zap.AddCaller()) // ← INI yang bikin tau file:line

	return logger, nil
}
