package logger

import (
	"github.com/sirupsen/logrus"
)

type Logger struct {
	flag  bool
	proto string
}

func New(flag bool, proto string) *Logger {
	logrus.SetLevel(logrus.DebugLevel)
	return &Logger{
		flag:  flag,
		proto: proto,
	}
}

func (l *Logger) DebugMode() bool {
	return l.flag
}

func (l *Logger) Info(args ...interface{}) {
	if l.flag {
		logrus.WithFields(logrus.Fields{
			"protocol": l.proto,
		}).Info(args)
	}
}

func (l *Logger) Debug(args ...interface{}) {
	if l.flag {
		logrus.WithFields(logrus.Fields{
			"protocol": l.proto,
		}).Debug(args)
	}
}

func (l *Logger) Warn(args ...interface{}) {
	if l.flag {
		logrus.WithFields(logrus.Fields{
			"protocol": l.proto,
		}).Warn(args)
	}
}

func (l *Logger) Error(args ...interface{}) {
	if l.flag {
		logrus.WithFields(logrus.Fields{
			"protocol": l.proto,
		}).Error(args)
	}
}

func (l *Logger) Infof(format string, args ...interface{}) {
	if l.flag {
		logrus.WithFields(logrus.Fields{
			"protocol": l.proto,
		}).Infof(format, args)
	}
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	if l.flag {
		logrus.WithFields(logrus.Fields{
			"protocol": l.proto,
		}).Debug(format, args)
	}
}

func (l *Logger) Warnf(foramt string, args ...interface{}) {
	if l.flag {
		logrus.WithFields(logrus.Fields{
			"protocol": l.proto,
		}).Warnf(foramt, args)
	}
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	if l.flag {
		logrus.WithFields(logrus.Fields{
			"protocol": l.proto,
		}).Errorf(format, args)
	}
}
