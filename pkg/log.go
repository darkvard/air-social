package pkg

import (
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	l    *zap.SugaredLogger
	once sync.Once
)

func NewLogger(env string) {
	once.Do(func() {
		l = newZapLogger(env)
	})
}

func Log() *zap.SugaredLogger {
	return l
}

func newZapLogger(env string) *zap.SugaredLogger {
	var cfg zapcore.EncoderConfig
	if env == "production" {
		cfg = zap.NewProductionEncoderConfig()
	} else {
		cfg = newEncoderColorConfig()
	}

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(cfg),
		zapcore.AddSync(os.Stdout),
		zap.InfoLevel,
	)

	logger := zap.New(core, zap.AddCaller())
	return logger.Sugar()
}

func newEncoderColorConfig() zapcore.EncoderConfig {
	cfg := zap.NewDevelopmentEncoderConfig()
	cfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	loc, _ := time.LoadLocation("Asia/Ho_Chi_Minh")
	cfg.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString("\033[35m" + t.In(loc).Format("2006-01-02 15:04:05") + "\033[0m")
	}
	cfg.EncodeCaller = func(c zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString("\033[90m" + c.TrimmedPath() + "\033[0m")
	}
	return cfg
}
