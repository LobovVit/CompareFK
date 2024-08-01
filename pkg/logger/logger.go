package logger

import (
	"fmt"
	"io"
	"time"

	"Compare/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	_ "gopkg.in/natefinch/lumberjack.v2"
)

var Log = zap.NewNop()

type WriteSyncer struct {
	io.Writer
}

func (ws WriteSyncer) Sync() error {
	return nil
}

func Initialize(config *config.Config) error {
	lvl, err := zap.ParseAtomicLevel(config.LogLevel)
	if err != nil {
		return fmt.Errorf("log parse level: %w", err)
	}
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)
	cfg.Encoding = "console"
	cfg.OutputPaths = []string{config.LogFile}
	sw := getWriteSyncer(config.LogFile)

	cfg.Level = lvl

	zl, err := cfg.Build(SetOutput(sw, cfg))
	if err != nil {
		return fmt.Errorf("log build: %w", err)
	}

	Log = zl
	return nil
}

func SetOutput(ws zapcore.WriteSyncer, conf zap.Config) zap.Option {
	var enc zapcore.Encoder
	switch conf.Encoding {
	case "json":
		enc = zapcore.NewJSONEncoder(conf.EncoderConfig)
	case "console":
		enc = zapcore.NewConsoleEncoder(conf.EncoderConfig)
	default:
		panic("unknown encoding")
	}

	return zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return zapcore.NewCore(enc, ws, conf.Level)
	})
}

func getWriteSyncer(logName string) zapcore.WriteSyncer {
	var ioWriter = &lumberjack.Logger{
		Filename:   logName,
		MaxSize:    10, // MB
		MaxBackups: 3,  // number of backups
		MaxAge:     28, //days
		LocalTime:  true,
		Compress:   false, // disabled by default
	}
	var sw = WriteSyncer{
		ioWriter,
	}
	return sw
}
