package log

import (
	"io"
	"log"
	"os"
	"fmt"
	"strings"
)

// Logger is the interface for logging messages.
type Logger interface {
	// Print writes a message to the log.
	Print(v ...interface{})
	// Printf writes a formatted message to the log.
	Printf(format string, v ...interface{})
	// Println writes a line to the log.
	Println(v ...interface{})
}

// Level represents the level of logging.
type Level int

func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarningLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	case DisabledLevel:
		return ""
	default:
		return "UNKNOWN"
	}
}

// Different levels of logging.
const (
	DebugLevel Level = iota
	InfoLevel
	WarningLevel
	ErrorLevel
	DisabledLevel
)

type logger struct {
	level Level
}

// Default loggers for each level.
var (
	Debug = &logger{DebugLevel}
	Info  = &logger{InfoLevel}
	Warn  = &logger{WarningLevel}
	Error = &logger{ErrorLevel}
)

var (
	currentLevel         = InfoLevel
	defaultLogger Logger = newDefaultLogger(os.Stderr)
)

func GetLevel() Level {
	return currentLevel
}

// SetLevel sets the current logging level.
func SetLevel(level Level) {
	currentLevel = level
}

func newDefaultLogger(w io.Writer) Logger {
	return log.New(w, "[DSA] ", log.Ldate | log.Lmicroseconds)
}

// SetOutput sets the default loggers to write to w.
// If w is nil, then default loggers are disabled.
func SetOutput(w io.Writer) {
	if w == nil {
		defaultLogger = nil
	} else {
		defaultLogger = newDefaultLogger(w)
	}
}

// Printf writes a formatted message to the log.
func (l *logger) Printf(format string, v ...interface{}) {
	if l.level < currentLevel {
		return
	}
	if defaultLogger != nil {
		f := appendLevel(format, l.level)
		defaultLogger.Printf(f, v...)
	}
}

func (l *logger) Print(v ...interface{}) {
	if l.level < currentLevel {
		return
	}
	if defaultLogger != nil {
		lev := appendLevel("", l.level)
		v := append([]interface{}{lev}, v...)
		defaultLogger.Print(v...)
	}
}

func (l *logger) Println(v ...interface{}) {
	if l.level < currentLevel {
		return
	}
	if defaultLogger != nil {
		lev := appendLevel("", l.level)
		v := append([]interface{}{lev}, v...)
		defaultLogger.Println(v...)
	}
}

func appendLevel(format string, l Level) string {
	return fmt.Sprintf("[%s] %s", l, format)
}

// At returns weather the level will be logged currently.
func At(level Level) bool {
	return currentLevel <= level
}

// ToLevel will attempt to convert string s to a valid
// Level. It will return Disabled and a non-nil error
// if it is an invalid string.
func ToLevel(s string) (Level, error) {
	switch strings.ToLower(s) {
	case "debug":
		return DebugLevel, nil
	case "info":
		return InfoLevel, nil
	case "warn", "warning":
		return WarningLevel, nil
	case "error":
		return ErrorLevel, nil
	case "disable", "disabled":
		return DisabledLevel, nil
	}
	return DisabledLevel, fmt.Errorf("invalid log level %q", s)
}

// Printf logs a formatted string at the INFO log level
func Printf(format string, v...interface{}) {
	Info.Printf(format, v...)
}

// Println logs a line at the INFO log level
func Println(v...interface{}) {
	Info.Println(v...)
}

// Print logs a string at the INFO log level
func Print(v...interface{}) {
	Info.Print(v...)
}