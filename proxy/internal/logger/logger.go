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
