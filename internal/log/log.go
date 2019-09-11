// Package log is a wrapper for external log packages
package log

import (
	"os"

	"github.com/sharenowTech/virity/internal/config"
	log "github.com/sirupsen/logrus"
)

var loglevel = map[string]log.Level{
	"DEBUG":    log.DebugLevel,
	"INFO":     log.InfoLevel,
	"WARN":     log.WarnLevel,
	"ERROR":    log.ErrorLevel,
	"CRITICAL": log.FatalLevel,
}

func init() {
	// Log as JSON instead of the default ASCII formatter.
	logtype := config.GetGeneralConfig().LogType
	switch logtype {
	case "JSON":
		log.SetFormatter(&log.JSONFormatter{})
	case "ASCII":
		log.SetFormatter(&log.TextFormatter{})
	default:
		Info(Fields{
			"package": "log",
			"logtype": logtype,
		}, "Logtype does not exist. Fallback to default logtype")

	}

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	if val, ok := loglevel[config.GetGeneralConfig().LogLevel]; ok {
		log.SetLevel(val)
	}
}

type Fields log.Fields

func Debug(f Fields, message string) {
	log.WithFields(log.Fields(f)).Debug(message)
}
func Info(f Fields, message string) {
	log.WithFields(log.Fields(f)).Info(message)
}
func Warn(f Fields, message string) {
	log.WithFields(log.Fields(f)).Warn(message)
}
func Error(f Fields, message string) {
	log.WithFields(log.Fields(f)).Error(message)
}
func Critical(f Fields, message string) {
	log.WithFields(log.Fields(f)).Fatal(message)
}

func Debugf(f Fields, format string, a ...interface{}) {
	log.WithFields(log.Fields(f)).Debugf(format, a...)
}
func Infof(f Fields, format string, a ...interface{}) {
	log.WithFields(log.Fields(f)).Infof(format, a...)
}
func Warnf(f Fields, format string, a ...interface{}) {
	log.WithFields(log.Fields(f)).Warnf(format, a...)
}
func Errorf(f Fields, format string, a ...interface{}) {
	log.WithFields(log.Fields(f)).Errorf(format, a...)
}
func Criticalf(f Fields, format string, a ...interface{}) {
	log.WithFields(log.Fields(f)).Fatalf(format, a...)
}
