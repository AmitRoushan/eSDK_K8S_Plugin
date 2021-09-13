package log

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	logFilePermission = 0644
	defaultFileSize   = 1024 * 1024 * 20
)

var (
	logFileSizeThreshold = flag.String("logFileSize",
		strconv.Itoa(defaultFileSize), // 20M
		"Maximum logging file size before truncation")
)

// FileHook sends log entries to a file.
type FileHook struct {
	logFilePath          string
	logFileHandle        *os.File
	logRotationThreshold int64
	formatter            logrus.Formatter
	mutex                *sync.Mutex
}

// NewFileHook creates a new log hook for writing to a file.
func NewFileHook(logFilePath string, logFormat logrus.Formatter) (*FileHook, error) {

	logRoot := filepath.Dir(logFilePath)
	dir, err := os.Lstat(logRoot)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(logRoot, 0755); err != nil {
			return nil, fmt.Errorf("could not create log directory %v. %v", logRoot, err)
		}
	}
	if dir != nil && !dir.IsDir() {
		return nil, fmt.Errorf("log path %v exists and is not a directory, please remove it", logRoot)
	}

	filesizeThreshold, err := getNumInByte()
	if err != nil {
		logrus.Errorf("Calc max log file size error: %v.", err)
		return nil, err
	}

	fileHandle, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_RDWR, logFilePermission)
	if err != nil {
		return nil, err
	}

	return &FileHook{
		logFilePath:          logFilePath,
		logRotationThreshold: filesizeThreshold,
		formatter:            logFormat,
		logFileHandle:        fileHandle,
		mutex:                &sync.Mutex{}}, nil
}

func (hook *FileHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (hook *FileHook) Fire(entry *logrus.Entry) error {
	// Get formatted entry
	lineBytes, err := hook.formatter.Format(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not read log entry. %v", err)
		return err
	}

	// Write log entry to file
	_, err = hook.logFileHandle.WriteString(string(lineBytes))
	if err != nil {
		logrus.Errorf("Write log message %s to %s error.", lineBytes, hook.logFilePath)
	}

	// Rotate the file as needed
	if err = hook.maybeDoLogfileRotation(); err != nil {
		return err
	}

	return nil
}

// logfileNeedsRotation checks to see if a file has grown too large
func (hook *FileHook) logfileNeedsRotation() bool {
	fileInfo, err := hook.logFileHandle.Stat()
	if err != nil {
		return false
	}

	return fileInfo.Size() >= hook.logRotationThreshold
}

// maybeDoLogfileRotation prevents descending into doLogfileRotation on every call as the inner
// func is somewhat expensive and doesn't really need to happen every log entry.
func (hook *FileHook) maybeDoLogfileRotation() error {
	// We use a mutex to protect rotation from concurrent loggers, but in order to avoid
	// contention over this resource with high logging levels, check the file before taking
	// the lock.  Only if the file needs rotating do we then acquire the lock and recheck
	// the size under it.  The winner of the lock race will rotate the file.
	if hook.logfileNeedsRotation() {
		hook.mutex.Lock()
		defer hook.mutex.Unlock()

		if hook.logfileNeedsRotation() {
			// Do the rotation.
			rotatedLogFileLocation := hook.logFilePath + time.Now().Format("20060102-150405")
			if err := os.Rename(hook.logFilePath, rotatedLogFileLocation); err != nil {
				return err
			}
		}
	}

	return nil
}

func getNumInByte() (int64, error) {
	var sum int64 = 0
	var err error

	maxDataNum := strings.ToUpper(*logFileSizeThreshold)
	lastLetter := maxDataNum[len(maxDataNum)-1:]

	// 1.最后一位是M
	// 1.1 获取M前面的数字 * 1024 * 1024
	// 2.最后一位是K
	// 2.1 获取K前面的数字 * 1024
	// 3.最后一位是数字或者B
	// 3.1 若最后一位是数字，则直接返回 若最后一位是B，则获取前面的数字返回
	if lastLetter >= "0" && lastLetter <= "9" {
		sum, err = strconv.ParseInt(maxDataNum, 10, 64)
		if err != nil {
			return 0, err
		}
	} else {
		sum, err = strconv.ParseInt(maxDataNum[:len(maxDataNum)-1], 10, 64)
		if err != nil {
			return 0, err
		}

		if lastLetter == "M" {
			sum *= 1024 * 1024
		} else if lastLetter == "K" {
			sum *= 1024
		}
	}

	return sum, nil
}
