package log

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
)

var (
	logger LoggingInterface

	loggingModule = flag.String("loggingModule",
		"file",
		"Flag enable one of available logging module (file, stdout, stderr)")
	logLevel = flag.String("logLevel",
		"info",
		"Set logging level (debug, error, info, warning, fatal)")
	logFileDir = flag.String("logFileDir",
		defaultLogDir,
		"The flag to specify logging directory. The flag is only supported if logging module is file")
)

const (
	defaultLogDir   = "/var/log/huawei"
	timestampFormat = "2006-01-02 15:04:05.000000"
)

type loggerImpl struct {
	*logrus.Logger
	hooks     []logrus.Hook
	formatter logrus.Formatter
}

var _ LoggingInterface = &loggerImpl{}

func parseLogLevel() (logrus.Level, error) {
	switch *logLevel {
	case "debug":
		return logrus.DebugLevel, nil
	case "info":
		return logrus.InfoLevel, nil
	case "warning":
		return logrus.WarnLevel, nil
	case "error":
		return logrus.ErrorLevel, nil
	case "fatal":
		return logrus.FatalLevel, nil
	default:
		return logrus.FatalLevel, fmt.Errorf("invalid logging level [%v]", logLevel)
	}
}

// InitLogging configures logging. Logs are written both to a log file as well as stdout/stderr.
// Since logrus doesn't support multiple writers, each log stream is implemented as a hook.
func InitLogging(logName string) error {
	var tmpLogger loggerImpl

	// initialize logrus in wrapper
	tmpLogger.Logger = logrus.New()

	// No output except for the hooks
	tmpLogger.Logger.SetOutput(ioutil.Discard)

	// set logging level
	level, err := parseLogLevel()
	if err != nil {
		return err
	}
	tmpLogger.Logger.SetLevel(level)

	// initialize log formatter
	formatter := &PlainTextFormatter{TimestampFormat: timestampFormat, pid: os.Getpid()}

	hooks := make([]logrus.Hook, 0)
	switch *loggingModule {
	case "file":
		logFilePath := fmt.Sprintf("%s/%s.log", *logFileDir, logName)
		// Write to the log file
		logFileHook, err := NewFileHook(logFilePath, formatter)
		if err != nil {
			return fmt.Errorf("could not initialize logging to file: %v", err)
		}
		hooks = append(hooks, logFileHook)
	case "console":
		// Write to stdout/stderr
		logConsoleHook, err := NewConsoleHook(formatter)
		if err != nil {
			return fmt.Errorf("could not initialize logging to console: %v", err)
		}
		hooks = append(hooks, logConsoleHook)
	default:
		return fmt.Errorf("invalid logging level [%v]", loggingModule)
	}

	tmpLogger.hooks = hooks
	for _, hook := range tmpLogger.hooks {
		// initialize logrus with hooks
		tmpLogger.Logger.AddHook(hook)
	}

	logger = &tmpLogger
	return nil
}

// PlainTextFormatter is a formatter than does no coloring *and* does not insist on writing logs as key/value pairs.
type PlainTextFormatter struct {

	// TimestampFormat to use for display when a full timestamp is printed
	TimestampFormat string

	// process identity number
	pid int
}

var _ logrus.Formatter = &PlainTextFormatter{}

func (f *PlainTextFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	b := entry.Buffer
	if entry.Buffer == nil {
		b = &bytes.Buffer{}
	}

	_, _ = fmt.Fprintf(b, "%s %d %s%s\n", entry.Time.Format(f.TimestampFormat), f.pid, getLogLevel(entry.Level), entry.Message)

	return b.Bytes(), nil
}

func getLogLevel(level logrus.Level) string {
	switch level {
	case logrus.DebugLevel:
		return "[DEBUG]: "
	case logrus.InfoLevel:
		return "[INFO]: "
	case logrus.WarnLevel:
		return "[WARNING]: "
	case logrus.ErrorLevel:
		return "[ERROR]: "
	case logrus.FatalLevel:
		return "[FATAL]: "
	default:
		return "[UNKNOWN]: "
	}
}
