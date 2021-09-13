package log

import (
	"fmt"
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

// ConsoleHook sends log entries to stdout/stderr.
type ConsoleHook struct {
	formatter logrus.Formatter
}

// NewConsoleHook creates a new log hook for writing to stdout/stderr.
func NewConsoleHook(logFormat logrus.Formatter) (*ConsoleHook, error) {
	return &ConsoleHook{logFormat}, nil
}

func (hook *ConsoleHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (hook *ConsoleHook) Fire(entry *logrus.Entry) error {

	// Determine output stream
	var logWriter io.Writer
	switch entry.Level {
	case logrus.DebugLevel, logrus.InfoLevel, logrus.WarnLevel:
		logWriter = os.Stdout
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		logWriter = os.Stderr
	default:
		return fmt.Errorf("unknown log level: %v", entry.Level)
	}

	lineBytes, err := hook.formatter.Format(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to read entry, %v", err)
		return err
	}

	if _, err := logWriter.Write(lineBytes); err != nil {
		return err
	}

	return nil
}
