package logger

import (
	"os"
	"strings"

	"github.com/rs/zerolog"
)

func New(level string) (zerolog.Logger, error) {
	parsedLevel, err := zerolog.ParseLevel(strings.ToLower(level))
	if err != nil {
		return zerolog.Logger{}, err
	}

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger().Level(parsedLevel)
	return logger, nil
}

type UseCaseAdapter struct {
	logger *zerolog.Logger
}

func NewUseCaseAdapter(logger *zerolog.Logger) *UseCaseAdapter {
	return &UseCaseAdapter{logger: logger}
}

func (a *UseCaseAdapter) Error(msg string, err error) {
	if a == nil || a.logger == nil {
		return
	}

	a.logger.Error().Err(err).Msg(msg)
}
