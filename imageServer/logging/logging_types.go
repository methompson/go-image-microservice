package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

/****************************************************************************************
* LoggingError
****************************************************************************************/
type LoggingError struct{ ErrMsg string }

func (err LoggingError) Error() string { return err.ErrMsg }
func NewLoggingError(msg string) error { return LoggingError{msg} }

/****************************************************************************************
* LogData
****************************************************************************************/
type LogData interface {
	PrettyString() string
}

/****************************************************************************************
* RequestLogData
****************************************************************************************/
type RequestLogData struct {
	Timestamp    time.Time     `bson:"timestamp"`
	Type         string        `bson:"type"`
	ClientIP     string        `bson:"clientIP"`
	Method       string        `bson:"method"`
	Path         string        `bson:"path"`
	Protocol     string        `bson:"protocol"`
	StatusCode   int           `bson:"statusCode"`
	Latency      time.Duration `bson:"latency"`
	UserAgent    string        `bson:"userAgent"`
	ErrorMessage string        `bson:"errorMessage"`
}

func (rld RequestLogData) PrettyString() string {
	msg := fmt.Sprintf("%s | %s %s %s %d | %s | %s | \"%s\" | \"%s\"",
		rld.Timestamp.Format(time.RFC1123),
		rld.Protocol,
		rld.Method,
		rld.Path,
		rld.StatusCode,
		rld.ClientIP,
		rld.Latency,
		rld.UserAgent,
		rld.ErrorMessage,
	)

	return msg
}

/****************************************************************************************
* InfoLogData
****************************************************************************************/
type InfoLogData struct {
	Timestamp time.Time `bson:"timestamp"`
	Type      string    `bson:"type"`
	Message   string    `bson:"message"`
}

func (ild InfoLogData) PrettyString() string {
	msg := fmt.Sprintf("%s - [%s] \"%s\"",
		ild.Timestamp.Format(time.RFC1123),
		ild.Type,
		ild.Message,
	)

	return msg
}

/****************************************************************************************
* ImageLogger
****************************************************************************************/
type ImageLogger interface {
	AddRequestLog(log RequestLogData) error
	AddInfoLog(log InfoLogData) error
}

/****************************************************************************************
* FileLogger
****************************************************************************************/
type FileLogger struct {
	FilePath   string
	FileName   string
	FileHandle *os.File
}

func (fl *FileLogger) AddRequestLog(log RequestLogData) error {
	err := fl.WriteLog(log)
	return err
}

func (fl *FileLogger) AddInfoLog(log InfoLogData) error {
	err := fl.WriteLog(log)
	return err
}

func (fl *FileLogger) WriteLog(log LogData) error {
	if fl.FileHandle == nil {
		return NewLoggingError("fileHandle is nil (no file handle exists)")
	}

	_, err := fl.FileHandle.WriteString(log.PrettyString() + "\n")

	return err
}

func MakeNewFileLogger(path string, name string) (fl *FileLogger, err error) {
	fl = &FileLogger{
		FileName: name,
		FilePath: path,
	}

	fullPath := filepath.Join(path, name)

	var pathErr error

	if _, err := os.Stat(path); os.IsNotExist(err) {
		pathErr = os.MkdirAll(path, 0764)
	}

	// We return the FileLogger with FileHandle set to nil
	if pathErr != nil {
		// do something
		return fl, pathErr
	}

	handle, handleErr := os.OpenFile(fullPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if handleErr != nil {
		return fl, handleErr
	}

	fl.FileHandle = handle

	return fl, nil
}

/****************************************************************************************
* ConsoleLogger
****************************************************************************************/
type ConsoleLogger struct {
}

func (cl *ConsoleLogger) AddRequestLog(log RequestLogData) error {
	fmt.Println(log.PrettyString())
	return nil
}

func (cl *ConsoleLogger) AddInfoLog(log InfoLogData) error {
	fmt.Println(log.PrettyString())
	return nil
}
