package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/rs/zerolog"

	"github.com/leshachaplin/datalog/internal/apierror"
	"github.com/leshachaplin/datalog/internal/service"
)

type Handler struct {
	eventProcessor service.Event
	logger         zerolog.Logger
}

func NewHandler(eventProcessor service.Event, logger zerolog.Logger) *Handler {
	return &Handler{
		eventProcessor: eventProcessor,
		logger:         logger,
	}
}

func (h *Handler) error(err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	var apiErr apierror.Error
	if !errors.As(err, &apiErr) {
		apiErr = apierror.NewAPIError(err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(apiErr.StatusCode())
	if err = json.NewEncoder(w).Encode(apiErr); err != nil {
		h.logger.Error().Err(err).Send() // TODO improve logging
	}
}
