package main

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetFormatter(logFormatFunc(formatLogEntry))
	log.Info("Hello World")
}

type logFormatFunc func(*log.Entry) ([]byte, error)

func (fn logFormatFunc) Format(e *log.Entry) ([]byte, error) {
	return fn(e)
}

func formatLogEntry(e *log.Entry) ([]byte, error) {
	return []byte(fmt.Sprintf("%s %s - %s", e.Time.Format("2006-01-02 15:04:05"), strings.ToUpper(e.Level.String()), e.Message)), nil
}
