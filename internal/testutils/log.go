package testutils

import (
	"io"
	"log/slog"
)

func NewNoOpLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, nil))
}
