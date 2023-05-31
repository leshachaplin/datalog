package http

import (
	"bytes"
	"io"
	"net/http"
	"time"
)

func (h *Handler) Event(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		h.error(err, w)
		return
	}
	buf := bytes.NewBuffer(data)
	go h.eventProcessor.ProcessEvent(buf, getClientIP(r), time.Now())
	w.WriteHeader(http.StatusAccepted)
}
