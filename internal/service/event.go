package service

import (
	"bufio"
	"bytes"
	"encoding/json"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/leshachaplin/datalog/internal/domain"
)

type Event interface {
	ProcessEvent(buf *bytes.Buffer, clientIP string, serverTime time.Time)
}

func (s *Service) ProcessEvent(buf *bytes.Buffer, clientIP string, serverTime time.Time) {
	l := log.Logger.With().Str("Service", "ProcessEvent").Logger()
	scanner := bufio.NewScanner(buf)
	scanner.Split(bufio.ScanLines)

	batch := domain.EventBatch{
		Events: make([]domain.Event, 0),
	}
	for scanner.Scan() {
		event := &domain.Event{}
		data := scanner.Bytes()
		if err := json.Unmarshal(data, event); err != nil {
			l.Err(err).Str("raw_event", string(data)).Msg("Failed to decode event")
			continue
		}
		event.EnrichWith(clientIP, serverTime)
		batch.Events = append(batch.Events, *event)
	}
	if len(batch.Events) > 0 {
		batch.ID = batch.Events[0].DeviceID
		s.eventPool.Process(batch)
	}
}
