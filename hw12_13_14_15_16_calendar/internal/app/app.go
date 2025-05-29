package app

import (
	"context"

	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/storage" //nolint:depguard,nolintlint
)

type App struct { // TODO
}

type Logger interface { // TODO
}

func New(logger Logger, storage storage.Storage) *App {
	return &App{}
}

func (a *App) CreateEvent(ctx context.Context, id, title string) error {
	// TODO
	return nil
	// return a.storage.CreateEvent(storage.Event{ID: id, Title: title})
}

// TODO
