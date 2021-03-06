package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"strings"

	log "github.com/Sirupsen/logrus"
)

const (
	logLogrusRepo     = "github.com/Sirupsen/logrus"
	logFileDateFormat = "2006-01-02 15:04:05.000-0700"
)

type fileHook struct{}

func (hook fileHook) Fire(entry *log.Entry) (err error) {
	pc := make([]uintptr, 5)
	cnt := runtime.Callers(6, pc)

	for i := 0; i < cnt; i++ {
		fu := runtime.FuncForPC(pc[i] - 1)
		name := fu.Name()

		if !strings.Contains(name, logLogrusRepo) {
			file, line := fu.FileLine(pc[i] - 1)
			entry.Data["file"] = fmt.Sprintf("%v:%v", path.Base(file), line)
			break
		}
	}

	return
}

func (hook fileHook) Levels() []log.Level {
	return log.AllLevels
}

type logFormatter struct{}

func (formatter logFormatter) Format(entry *log.Entry) ([]byte, error) {
	var buffer bytes.Buffer
	file, _ := entry.Data["file"].(string)

	formatter.formatBracketEntry(&buffer, entry.Time.Format(logFileDateFormat))
	formatter.formatBracketEntry(&buffer, strings.ToUpper(entry.Level.String()))
	formatter.formatBracketEntry(&buffer, file)

	buffer.WriteString("- ")

	for _, char := range entry.Message {
		switch {
		case char == '\n':
			buffer.WriteString("\\n")
		case char < 32 || char == 127:
			buffer.WriteString(NonPrintableCharString(char))
		default:
			buffer.WriteRune(char)
		}
	}

	buffer.WriteRune('\n')

	return buffer.Bytes(), nil
}

func (formatter logFormatter) formatBracketEntry(buffer *bytes.Buffer, value string) {
	buffer.WriteRune('[')
	buffer.WriteString(value)
	buffer.WriteString("] ")
}

// InitialiseLogging sets up logging
func InitialiseLogging(logLevel, logFilePath string) {
	if logLevel == MnLogLevelDefault {
		log.SetOutput(ioutil.Discard)
		return
	}

	logLevels := map[string]log.Level{
		"PANIC": log.PanicLevel,
		"FATAL": log.FatalLevel,
		"ERROR": log.ErrorLevel,
		"WARN":  log.WarnLevel,
		"INFO":  log.InfoLevel,
		"DEBUG": log.DebugLevel,
	}

	if level, ok := logLevels[logLevel]; ok {
		log.SetLevel(level)
	} else {
		log.Fatalf("Invalid logLevel: %v", logLevel)
	}

	file, err := os.OpenFile(logFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatalf("Unable to open log file %v for writing: %v", logFilePath, err)
	}

	log.SetOutput(file)

	log.SetFormatter(logFormatter{})

	log.AddHook(fileHook{})
}
