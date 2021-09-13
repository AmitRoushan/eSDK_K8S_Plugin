package log

import "os"

type LoggingInterface interface {
	Debugf(format string, args ...interface{})

	Debugln(args ...interface{})

	Infof(format string, args ...interface{})

	Infoln(args ...interface{})

	Warningf(format string, args ...interface{})

	Warningln(args ...interface{})

	Errorf(format string, args ...interface{})

	Errorln(args ...interface{})

	Fatalf(format string, args ...interface{})

	Fatalln(args ...interface{})

	Flushable

	Closable
}

type Closable interface {
	Close()
}

type Flushable interface {
	Flush()
}

func Debugf(format string, args ...interface{}) {
	logger.Debugf(format, args...)
}

func Debugln(args ...interface{}) {
	logger.Debugln(args...)
}

func Infof(format string, args ...interface{}) {
	logger.Infof(format, args...)
}

func Infoln(args ...interface{}) {
	logger.Infoln(args...)
}

func Warningf(format string, args ...interface{}) {
	logger.Warningf(format, args...)
}

func Warningln(args ...interface{}) {
	logger.Warningln(args...)
}

func Errorf(format string, args ...interface{}) {
	logger.Errorf(format, args...)
}

func Errorln(args ...interface{}) {
	logger.Errorln(args...)
}

func Fatalf(format string, args ...interface{}) {
	logger.Fatalf(format, args...)
	os.Exit(255)
}

func Fatalln(args ...interface{}) {
	logger.Fatalln(args...)
	os.Exit(255)
}

func (logger *loggerImpl) Flush() {
	for _, hook := range logger.hooks {
		flushable, ok := hook.(Flushable)
		if ok {
			flushable.Flush()
		}
	}
}

func (logger *loggerImpl) Close() {
	for _, hook := range logger.hooks {
		flushable, ok := hook.(Closable)
		if ok {
			flushable.Close()
		}
	}
}

func Flush() {
	logger.Flush()
}

func Close() {
	logger.Close()
}
