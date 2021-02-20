package util

import (
	"fmt"
	"io"
	"log"
	"os"
)

type Level int

const (
	DEBUG Level = iota + 1
	INFO
	WARN
	ERROR
	FATAL
	PANIC
)

type Logger struct {
	*log.Logger
	level  Level
	prefix string
}

func init() {
	logLevel := os.Getenv("GOEX_LOG_LEVEL")
	if logLevel != "" {
		var l Level
		switch logLevel {
		case "debug", "DEBUG":
			l = DEBUG
		case "info", "INFO":
			l = INFO
		case "warn", "WARN":
			l = WARN
		case "error", "ERROR":
			l = ERROR
		case "fatal", "FATAL":
			l = FATAL
		case "panic", "PANIC":
			l = PANIC
		default:
			l = ERROR
		}
		SetLevel(l)
	}

	logFileName := os.Getenv("GOEX_LOG_FILE")
	if logFileName != "" {
		f, err := os.OpenFile(logFileName, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
		if err == nil {
			SetOut(f)
		} else {
			Warn("log file not open ??? ")
			Error(err.Error())
		}
	}

	logPrefix := os.Getenv("GOEX_LOG_PREFIX")
	if logPrefix != "" {
		SetPrefix(logPrefix)
	}
}

var Log = NewLogger()

func SetOut(out io.Writer) {
	Log.SetOut(out)
}

func SetLevel(level Level) {
	Log.SetLevel(level)
}

func SetPrefix(prefix string) {
	Log.SePrefix(prefix)
}

func Debug(args ...interface{}) {
	Log.Debug(args...)
}

func Debugf(format string, args ...interface{}) {
	Log.Debugf(format, args...)
}

func Info(args ...interface{}) {
	Log.Info(args...)
}

func Infof(format string, args ...interface{}) {
	Log.Infof(format, args...)
}

func Warn(args ...interface{}) {
	Log.Warn(args...)
}

func Warnf(format string, args ...interface{}) {
	Log.Warnf(format, args...)
}

func Error(args ...interface{}) {
	Log.Error(args...)
}

func Errorf(format string, args ...interface{}) {
	Log.Errorf(format, args...)
}

func Fatal(args ...interface{}) {
	Log.Fatal(args...)
}

func Fatalf(format string, args ...interface{}) {
	Log.Fatalf(format, args...)
}

func Panic(args ...interface{}) {
	Log.Panic(args...)
}

func Panicf(format string, args ...interface{}) {
	Log.Panicf(format, args...)
}

func NewLogger() *Logger {
	return &Logger{
		Logger: log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile),
		level:  INFO,
	}
}

func (l *Logger) SetLevel(level Level) {
	l.level = level
}

func (l *Logger) SetOut(out io.Writer) {
	l.Logger.SetOutput(out)
}

func (l *Logger) SePrefix(prefix string) {
	l.prefix = " " + prefix
}

func (l *Logger) output(le Level, prefix string, log string) {
	if l.level <= le {
		l.Output(3, fmt.Sprintf("%s %s", prefix, log))
	}
}

func (l *Logger) Debug(args ...interface{}) {
	l.output(DEBUG, "[D]"+l.prefix, fmt.Sprint(args...))
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.output(DEBUG, "[D]"+l.prefix, fmt.Sprintf(format, args...))
}

func (l *Logger) Info(args ...interface{}) {
	l.output(INFO, "[I]"+l.prefix, fmt.Sprint(args...))
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.output(INFO, "[I]"+l.prefix, fmt.Sprintf(format, args...))
}

func (l *Logger) Warn(args ...interface{}) {
	l.output(WARN, "[W]"+l.prefix, fmt.Sprint(args...))
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	l.output(WARN, "[W]"+l.prefix, fmt.Sprintf(format, args...))
}

func (l *Logger) Error(args ...interface{}) {
	l.output(ERROR, "[E]"+l.prefix, fmt.Sprint(args...))
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.output(ERROR, "[E]"+l.prefix, fmt.Sprintf(format, args...))
}

func (l *Logger) Fatal(args ...interface{}) {
	if l.level <= FATAL {
		l.output(FATAL, "[F]"+l.prefix, fmt.Sprint(args...))
		os.Exit(1)
	}
}

func (l *Logger) Fatalf(format string, args ...interface{}) {
	if l.level <= FATAL {
		l.output(FATAL, "[F]"+l.prefix, fmt.Sprintf(format, args...))
		os.Exit(1)
	}
}

func (l *Logger) Panic(args ...interface{}) {
	if l.level <= PANIC {
		s := fmt.Sprint(args...)
		l.output(PANIC, "[P]"+l.prefix, s)
		panic(s)
	}
}

func (l *Logger) Panicf(format string, args ...interface{}) {
	if l.level <= PANIC {
		s := fmt.Sprintf(format, args...)
		l.output(PANIC, "[P]"+l.prefix, s)
		panic(s)
	}
}
