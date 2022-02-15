package log

import (
	"io"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	zap    *zap.SugaredLogger
	config *Config
	fields []interface{}

	// logFile *os.File
}

type Config struct {
	Level string `default:"info"`
	File  struct {
		Enabled bool   `default:"true"`
		Path    string `default:"/var/log/tergum.log"`
	}
	TTY             bool   `default:"false"`
	TimestampFormat string `default:"2006-01-02 15:04:05"`
}

var zapConfig = zapcore.EncoderConfig{
	TimeKey:        "ts",
	LevelKey:       "level",
	NameKey:        "logger",
	CallerKey:      "caller",
	FunctionKey:    zapcore.OmitKey,
	MessageKey:     "msg",
	StacktraceKey:  "stacktrace",
	LineEnding:     zapcore.DefaultLineEnding,
	EncodeLevel:    zapcore.LowercaseLevelEncoder,
	EncodeTime:     zapcore.EpochTimeEncoder,
	EncodeDuration: zapcore.SecondsDurationEncoder,
	EncodeCaller:   zapcore.ShortCallerEncoder,
}

func newZapLogger(conf *Config) (*zap.SugaredLogger, error) {
	level, err := zap.ParseAtomicLevel(conf.Level)
	if err != nil {
		return nil, err
	}

	consoleOut := zapcore.Lock(os.Stdout)

	fileOut := zapcore.AddSync(io.Discard)
	if conf.File.Enabled {
		file, err := os.OpenFile(conf.File.Path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return nil, err
		}

		fileOut = zapcore.Lock(file)
	}

	newZapConf := zapConfig
	newZapConf.EncodeTime = zapcore.TimeEncoderOfLayout(conf.TimestampFormat)

	fileEncoder := zapcore.NewJSONEncoder(newZapConf)
	consoleEncoder := zapcore.NewConsoleEncoder(newZapConf)

	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, consoleOut, level),
		zapcore.NewCore(fileEncoder, fileOut, level),
	)

	return zap.New(core).Sugar(), nil
}

func New(conf *Config, fields ...interface{}) (*Logger, error) {
	zapLogger, err := newZapLogger(conf)
	if err != nil {
		return nil, err
	}

	logger := &Logger{
		zap:    zapLogger,
		config: conf,
		fields: fields,
	}

	return logger, nil
}

func (log *Logger) Close() error {
	return log.zap.Sync()
}

func (log *Logger) WithFields(fields ...interface{}) *Logger {
	return &Logger{
		zap:    log.zap,
		config: log.config,
		fields: fields,
	}
}

func (log *Logger) GetLevel() string {
	return log.config.Level
}

func (log *Logger) SetLevel(level string) bool {
	prevLevel := log.config.Level
	log.config.Level = level

	zapLogger, err := newZapLogger(log.config)
	if err != nil {
		log.config.Level = prevLevel
		return false
	}

	log.zap = zapLogger
	return true
}

func (log *Logger) Panic(msg ...interface{}) {
	if log.fields != nil && len(log.fields) != 0 {
		log.zap.With(log.fields...).Panic(msg...)
		return
	}
	log.zap.Panic(msg...)
}

func (log *Logger) Fatal(msg ...interface{}) {
	if log.fields != nil && len(log.fields) != 0 {
		log.zap.With(log.fields...).Fatal(msg...)
		return
	}
	log.zap.Fatal(msg...)
}

func (log *Logger) Debug(msg ...interface{}) {
	log.zap.With(log.fields...).Debug(msg...)
}

func (log *Logger) Info(msg ...interface{}) {
	if log.fields != nil && len(log.fields) != 0 {
		log.zap.With(log.fields...).Info(msg...)
		return
	}
	log.zap.Info(msg...)
}

func (log *Logger) Warn(msg ...interface{}) {
	if log.fields != nil && len(log.fields) != 0 {
		log.zap.With(log.fields...).Warn(msg...)
		return
	}
	log.zap.Warn(msg...)
}

func (log *Logger) Error(msg ...interface{}) {
	if log.fields != nil && len(log.fields) != 0 {
		log.zap.With(log.fields...).Error(msg...)
		return
	}
	log.zap.Error(msg...)
}
