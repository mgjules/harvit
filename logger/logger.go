package logger

import (
	"fmt"
	"io"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
)

// Logger is a simple wrapper around zap.SugaredLogger.
type Logger struct {
	*otelzap.SugaredLogger
}

// New creates a new Logger.
func New(prod bool) (*Logger, error) {
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

	return &Logger{otellogger.Sugar()}, nil
}

// Writer returns the logger's io.Writer.
func (l *Logger) Writer() io.Writer {
	return zap.NewStdLog(l.Desugar().Logger).Writer()
}
