package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"chainfeed-go/internal/config"
)

func New(cfg config.LogConfig) (*zap.Logger, error) {
	level := zapcore.InfoLevel
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		return nil, err
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	var encoder zapcore.Encoder
	if cfg.Encoding == "console" {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	var cores []zapcore.Core
	for _, path := range cfg.OutputPaths {
		var writer zapcore.WriteSyncer
		if path == "stdout" {
			writer = zapcore.AddSync(os.Stdout)
		} else {
			file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return nil, err
			}
			writer = zapcore.AddSync(file)
		}
		cores = append(cores, zapcore.NewCore(encoder, writer, level))
	}

	core := zapcore.NewTee(cores...)
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return logger, nil
}
