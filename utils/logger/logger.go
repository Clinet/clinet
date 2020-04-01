package logger

/*
	verbosity := 2 (0 = default, 1 = debug, 2 = trace)
	log := logger.NewLogger(verbosity)
	log.Trace("Something very low level.")
	log.Debug("Useful debugging information.")
	log.Info("Something noteworthy happened!")
	log.Warn("You should probably take a look at this.")
	log.Error("Something failed but I'm not quitting.")
	// Calls os.Exit(1) after logging
	log.Fatal("Bye.")
	// Calls panic() after logging
	log.Panic("I'm bailing.")
*/

import (
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

type Logger struct {
	logger *logrus.Logger
	prefix string
}

//NewLogger returns a logger with the specified verbosity level.
// prefix: string: the prefix for the logger
// verbosity: int: declares the verbosity level
//  - 0: default logging (info, warning, error)
//  - 1: includes 0, plus debug logging
//  - 2: includes 1, plus trace logging
func NewLogger(prefix string, verbosity int) *Logger {
	formatter := new(prefixed.TextFormatter)
	formatter.FullTimestamp = true

	log := logrus.New()
	log.Formatter = formatter

	switch verbosity {
	case 2:
		log.Level = logrus.TraceLevel
	case 1:
		log.Level = logrus.DebugLevel
	case 0:
		log.Level = logrus.InfoLevel
	}

	return &Logger{
		logger: log,
		prefix: prefix,
	}
}

func (logger *Logger) Trace(args ...interface{}) {
	logger.logger.WithField("prefix", logger.prefix).Trace(args...)
}
func (logger *Logger) Debug(args ...interface{}) {
	logger.logger.WithField("prefix", logger.prefix).Debug(args...)
}
func (logger *Logger) Info(args ...interface{}) {
	logger.logger.WithField("prefix", logger.prefix).Info(args...)
}
func (logger *Logger) Warn(args ...interface{}) {
	logger.logger.WithField("prefix", logger.prefix).Warn(args...)
}
func (logger *Logger) Error(args ...interface{}) {
	logger.logger.WithField("prefix", logger.prefix).Error(args...)
}
func (logger *Logger) Fatal(args ...interface{}) {
	logger.logger.WithField("prefix", logger.prefix).Fatal(args...)
}
func (logger *Logger) Panic(args ...interface{}) {
	logger.logger.WithField("prefix", logger.prefix).Panic(args...)
}