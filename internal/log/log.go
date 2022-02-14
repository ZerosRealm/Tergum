package log

import (
	"os"
	"sync"

	"github.com/sirupsen/logrus"
)

type Logger struct {
	stdOut  *logrus.Logger
	fileOut *logrus.Logger

	mutex  sync.RWMutex
	fields logrus.Fields

	logFile *os.File
}

type Config struct {
	Level           string `default:"info"`
	File            string `default:"/var/log/tergum.log"`
	TimestampFormat string `default:"2006-01-02 15:04:05"`
	TTY             bool   `default:"false"`
}

func New(conf *Config, fields map[string]interface{}) (*Logger, error) {
	stdLogger := logrus.New()
	stdLogger.SetFormatter(&logrus.TextFormatter{
		DisableColors:   false,
		FullTimestamp:   true,
		TimestampFormat: conf.TimestampFormat,
		ForceColors:     conf.TTY,
	})

	level, err := logrus.ParseLevel(conf.Level)
	if err != nil {
		return nil, err
	}
	stdLogger.SetLevel(level)

	// if fields == nil {
	// 	fields = make(map[string]interface{})
	// }

	logger := &Logger{
		stdOut: stdLogger,
		fields: fields,
		mutex:  sync.RWMutex{},
	}

	if conf.File != "" {
		file, err := os.OpenFile(conf.File, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return nil, err
		}

		fileLogger := logrus.New()
		fileLogger.SetFormatter(&logrus.JSONFormatter{
			DisableTimestamp: false,
			TimestampFormat:  conf.TimestampFormat,
		})
		fileLogger.SetLevel(level)
		fileLogger.SetOutput(file)

		logger.fileOut = fileLogger
		logger.logFile = file
	}

	return logger, nil
}

func (log *Logger) Close() error {
	if log.logFile != nil {
		return nil
	}

	return log.logFile.Close()
}

func (log *Logger) WithFields(fields ...interface{}) *Logger {
	log.mutex.Lock()
	defer log.mutex.Unlock()

	newFields := log.fields
	if newFields == nil {
		newFields = make(map[string]interface{})
	}

	for i := 0; i < len(fields)-1; i += 2 {
		switch v := fields[i].(type) {
		case string:
			newFields[v] = fields[i+1]
		default:
			panic("logger.WithFields: field name must be string")
		}
	}

	// Copy the logger, including mutex to prevent consurrent map iteration and map write.
	return &Logger{
		stdOut:  log.stdOut,
		fileOut: log.fileOut,
		fields:  newFields,
		mutex:   log.mutex,
	}
}

func (log *Logger) Panic(msg ...interface{}) {
	log.mutex.Lock()
	defer log.mutex.Unlock()

	log.stdOut.WithFields(log.fields).Panicln(msg...)
	if log.fileOut != nil {
		log.fileOut.WithFields(log.fields).Panicln(msg...)
	}
}

func (log *Logger) Fatal(msg ...interface{}) {
	log.mutex.Lock()
	defer log.mutex.Unlock()

	log.stdOut.WithFields(log.fields).Fatalln(msg...)
	if log.fileOut != nil {
		log.fileOut.WithFields(log.fields).Fatalln(msg...)
	}
}

func (log *Logger) Trace(msg ...interface{}) {
	log.mutex.Lock()
	defer log.mutex.Unlock()

	log.stdOut.WithFields(log.fields).Traceln(msg...)
	if log.fileOut != nil {
		log.fileOut.WithFields(log.fields).Traceln(msg...)
	}
}

func (log *Logger) Debug(msg ...interface{}) {
	log.mutex.Lock()
	defer log.mutex.Unlock()

	log.stdOut.WithFields(log.fields).Debugln(msg...)
	if log.fileOut != nil {
		log.fileOut.WithFields(log.fields).Debugln(msg...)
	}
}

func (log *Logger) Warn(msg ...interface{}) {
	log.mutex.Lock()
	defer log.mutex.Unlock()

	log.stdOut.WithFields(log.fields).Warnln(msg...)
	if log.fileOut != nil {
		log.fileOut.WithFields(log.fields).Warnln(msg...)
	}
}

func (log *Logger) Info(msg ...interface{}) {
	log.mutex.Lock()
	defer log.mutex.Unlock()

	log.stdOut.WithFields(log.fields).Infoln(msg...)
	if log.fileOut != nil {
		log.fileOut.WithFields(log.fields).Infoln(msg...)
	}
}

func (log *Logger) Error(msg ...interface{}) {
	log.mutex.Lock()
	defer log.mutex.Unlock()

	log.stdOut.WithFields(log.fields).Errorln(msg...)
	if log.fileOut != nil {
		log.fileOut.WithFields(log.fields).Errorln(msg...)
	}
}
