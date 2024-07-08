package log

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

type Logger struct {
	zlog zerolog.Logger
}

func NewLogger(level string) *Logger {
	logLevel, err := zerolog.ParseLevel(strings.ToLower(level))
	if err != nil {
		fmt.Printf("Invalid log level '%s', defaulting to 'info'\n", level)
		logLevel = zerolog.InfoLevel
	}

	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}

	zlog := zerolog.New(output).
		Level(logLevel).
		With().
		Timestamp().
		Logger()

	return &Logger{zlog: zlog}
}

func (l *Logger) Trace(msg string, args ...interface{}) {
	l.zlog.Trace().Msgf(msg, args...)
}

func (l *Logger) Debug(msg string, args ...interface{}) {
	l.zlog.Debug().Msgf(msg, args...)
}

func (l *Logger) Info(msg string, args ...interface{}) {
	l.zlog.Info().Msgf(msg, args...)
}

func (l *Logger) Warn(msg string, args ...interface{}) {
	l.zlog.Warn().Msgf(msg, args...)
}

func (l *Logger) Warning(msg string, args ...interface{}) {
	l.zlog.Warn().Msgf(msg, args...)
}

func (l *Logger) Error(msg string, args ...interface{}) {
	l.zlog.Error().Msgf(msg, args...)
}

func (l *Logger) Fatal(msg string, args ...interface{}) {
	l.zlog.Fatal().Msgf(msg, args...)
}

func (l *Logger) Panel(msg string, args ...interface{}) {
	l.zlog.Info().Str("source", "panel").Msgf(msg, args...)
}
