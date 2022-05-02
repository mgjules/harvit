package logger

import (
	"fmt"
	"io"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
)

// Log is a global instance of Logger
var Log *Logger

// Logger is a simple wrapper around zap.SugaredLogger.
type Logger struct {
	*otelzap.SugaredLogger
}

// New creates a new Logger.
func New(prod bool) (*Logger, error) {
	if Log != nil {
		return Log, nil
	}

	var (
		logger *zap.Logger
		err    error
	)

	if prod {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	otellogger := otelzap.New(logger)

	Log = &Logger{otellogger.Sugar()}

	return Log, nil
}

// Writer returns the logger's io.Writer.
func (l *Logger) Writer() io.Writer {
	return zap.NewStdLog(l.Desugar().Logger).Writer()
}
