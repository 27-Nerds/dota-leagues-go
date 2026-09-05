package handler

import (
	"context"
	e "dota_league/error"
	"dota_league/model"
)

type UpdatesReader interface {
	GetAll(context.Context, int, int, ...model.UpdateFilter) ([]model.Update, int64, error)
}

type UpdatesHandler struct{ repository UpdatesReader }

func NewUpdatesHandler(repository UpdatesReader) *UpdatesHandler {
	return &UpdatesHandler{repository: repository}
}

func (h *UpdatesHandler) GetAll(ctx context.Context, offset, limit int, filters ...model.UpdateFilter) ([]model.Update, int64, error) {
	rows, total, err := h.repository.GetAll(ctx, offset, limit, filters...)
	if err != nil {
		return nil, 0, &e.Error{Op: "UpdatesHandler.GetAll", Err: err}
	}
	return rows, total, nil
}
