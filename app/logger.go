package app

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

type Level string

const (
	TRACE Level = "TRACE"
	DEBUG Level = "DEBUG"
	INFO  Level = "INFO"
	WARN  Level = "WARN"
	ERROR Level = "ERROR"
	PANIC Level = "PANIC"
)

func NewZeroLogger(logLevel Level) zerolog.Logger {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	return zerolog.New(os.Stdout).
		Level(logLevelToZero(logLevel)).
		With().
		Timestamp().
		Caller().
		Logger()
}

func logLevelToZero(level Level) zerolog.Level {
	switch level {
	case PANIC:
		return zerolog.PanicLevel
	case ERROR:
		return zerolog.ErrorLevel
	case WARN:
		return zerolog.WarnLevel
	case INFO:
		return zerolog.InfoLevel
	case DEBUG:
		return zerolog.DebugLevel
	case TRACE:
		return zerolog.TraceLevel
	default:
		return zerolog.InfoLevel
	}
}
