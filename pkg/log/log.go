package log

import (
	"fmt"
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger interface {
	Debugf(template string, args ...interface{})
	Infof(template string, args ...interface{})
	Warnf(template string, args ...interface{})
	Errorf(template string, args ...interface{})
	Panicf(template string, args ...interface{})
	Fatalf(template string, args ...interface{})

	Debugw(msg string, keyAndValues ...interface{})
	Infow(msg string, keyAndValues ...interface{})
	Warnw(msg string, keyAndValues ...interface{})
	Errorw(msg string, keyAndValues ...interface{})
	Panicw(msg string, keyAndValues ...interface{})
	Fatalw(msg string, keysAndValues ...interface{})

	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})

	Sync()
}

var defaultLogger Logger = &nullLogger{}

var once sync.Once

func GetLogger(config *Config) (Logger, error) {
	var err error
	once.Do(func() {
		if config == nil {
			err = fmt.Errorf("config is nil")
			return
		}
		level := zap.InfoLevel
		if err = level.Set(config.Level); err != nil {
			return
		}
		var encoder zapcore.Encoder
		switch config.Encoding {
		case "json":
			encodeCfg := zap.NewProductionEncoderConfig()
			encodeCfg.EncodeTime = zapcore.ISO8601TimeEncoder
			encoder = zapcore.NewJSONEncoder(encodeCfg)
		case "console":
			encoder = zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
		default:
			err = fmt.Errorf("unknown encoding %s", config.Encoding)
			return
		}

		core := zapcore.NewCore(
			encoder,
			zapcore.NewMultiWriteSyncer(makeSyncers(config.OutputPaths)...),
			zap.LevelEnablerFunc(func(lvl zapcore.Level) bool { return lvl >= level }),
		)

		logger := zap.New(core)

		defaultLogger = &zapLogger{
			Logger: logger,
		}
	})
	return defaultLogger, err
}

func makeSyncers(paths []string) []zapcore.WriteSyncer {
	syncers := make([]zapcore.WriteSyncer, len(paths))
	for i, path := range paths {
		switch path {
		case "stdout":
			syncers[i] = zapcore.Lock(os.Stdout)
		case "stderr":
			syncers[i] = zapcore.Lock(os.Stderr)
		default:
			syncers[i] = zapcore.AddSync(&lumberjack.Logger{
				Filename: path,
			})
		}
	}
	return syncers
}

type zapLogger struct {
	*zap.Logger
}

// Debugf ...
func (z *zapLogger) Debugf(template string, args ...interface{}) {
	z.Sugar().Debugf(template, args...)
}

// Infof ...
func (z *zapLogger) Infof(template string, args ...interface{}) {
	z.Sugar().Infof(template, args...)
}

// Warnf ...
func (z *zapLogger) Warnf(template string, args ...interface{}) {
	z.Sugar().Warnf(template, args...)
}

// Errorf ...
func (z *zapLogger) Errorf(template string, args ...interface{}) {
	z.Sugar().Errorf(template, args...)
}

// Panicf ...
func (z *zapLogger) Panicf(template string, args ...interface{}) {
	z.Sugar().Panicf(template, args...)
}

// Fatalf ...
func (z *zapLogger) Fatalf(template string, args ...interface{}) {
	z.Sugar().Fatalf(template, args...)
}

// Debugw ...
func (z *zapLogger) Debugw(msg string, keysAndValues ...interface{}) {
	z.Sugar().Debugw(msg, keysAndValues...)
}

// Infow ...
func (z *zapLogger) Infow(msg string, keysAndValues ...interface{}) {
	z.Sugar().Infow(msg, keysAndValues...)
}

// Warnw ...
func (z *zapLogger) Warnw(msg string, keysAndValues ...interface{}) {
	z.Sugar().Warnw(msg, keysAndValues...)
}

// Errorw ...
func (z *zapLogger) Errorw(msg string, keysAndValues ...interface{}) {
	z.Sugar().Errorw(msg, keysAndValues...)
}

// Panicw ...
func (z *zapLogger) Panicw(msg string, keysAndValues ...interface{}) {
	z.Sugar().Panicw(msg, keysAndValues...)
}

// Fatalw ...
func (z *zapLogger) Fatalw(msg string, keysAndValues ...interface{}) {
	z.Sugar().Fatalw(msg, keysAndValues...)
}

func (z *zapLogger) Debug(args ...interface{}) {
	z.Sugar().Debug(args...)
}

func (z *zapLogger) Info(args ...interface{}) {
	z.Sugar().Info(args...)
}

func (z *zapLogger) Warn(args ...interface{}) {
	z.Sugar().Warn(args...)
}

func (z *zapLogger) Error(args ...interface{}) {
	z.Sugar().Error(args...)
}

func (z *zapLogger) Fatal(args ...interface{}) {
	z.Sugar().Fatal(args...)
}

// Sync ...
func (z *zapLogger) Sync() {
	_ = z.Sugar().Sync()
}

func Close() {
	defaultLogger.Sync()
}
