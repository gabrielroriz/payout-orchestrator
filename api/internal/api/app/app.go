package app

import "log/slog"

type Application struct {
	Logger *slog.Logger
}

func NewApplication(logger *slog.Logger) (*Application, error) {
	return &Application{logger}, nil
}
