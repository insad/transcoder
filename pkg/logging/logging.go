package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"logur.dev/logur"
)

var (
	loggers     = map[string]*zap.SugaredLogger{}
	environment = EnvDebug

	EnvDebug = "debug"
	EnvProd  = "prod"
)

var Prod = zap.NewProductionConfig()
var Dev = zap.NewDevelopmentConfig()

func init() {
	Prod.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
}

func Create(name string, cfg zap.Config) *zap.SugaredLogger {
	l, _ := cfg.Build()
	return l.Named(name).Sugar()
}

type KVLogger interface {
	Debug(msg string, keyvals ...interface{})
	Info(msg string, keyvals ...interface{})
	Warn(msg string, keyvals ...interface{})
	Error(msg string, keyvals ...interface{})
	With(keyvals ...interface{}) KVLogger
}

type NoopKVLogger struct {
	logur.NoopKVLogger
}

func (l NoopKVLogger) With(keyvals ...interface{}) KVLogger {
	return l
}
