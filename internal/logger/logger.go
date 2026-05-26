package logger

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.SugaredLogger
}

func NewLogger(env string) *Logger {
	cfg := zap.NewDevelopmentConfig()

	if env == "production" {
		cfg = zap.NewProductionConfig()
		cfg.Encoding = "json"
	} else {
		// 👇 culori + format dev friendly
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

		// 👇 human readable time
		cfg.EncoderConfig.EncodeTime = humanTimeEncoder
	}

	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}

	return &Logger{
		SugaredLogger: logger.Sugar(),
	}
}

// 👇 custom time format
func humanTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("02-01-2006 15:04:05"))
}
