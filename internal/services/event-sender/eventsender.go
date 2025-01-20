package eventsender

import (
	"context"
	"log/slog"
	"time"

	"URL-shortener/internal/domain"
	"URL-shortener/internal/storage/sqlite"
)

type Sender struct {
	storage *sqlite.Storage
	log     *slog.Logger
}

func NewSender(storage *sqlite.Storage, log *slog.Logger) *Sender {
	return &Sender{
		storage: storage,
		log:     log,
	}
}

func (s *Sender) StartProcessingEvents(ctx context.Context, handlePeriod time.Duration) {
	const op = "services.eventsender.StartProcessingEvents"

	log := s.log.With(slog.String("op", op))

	ticker := time.NewTicker(handlePeriod)

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Info("stop event processing")
				return
			case <-ticker.C:
			}

			evt, err := s.storage.GetNewEvent()
			if err != nil {
				log.Error("failed to get event", op, err)
				continue
			}
			if evt.ID == 0 {
				log.Debug("no new events")
				continue
			}

			s.SendMessage(evt)

			if err = s.storage.SetEventDone(evt.ID); err != nil {
				log.Error("failed to set event done", op, err)
				continue
			}
		}
	}()
}

func (s *Sender) SendMessage(event domain.Event) {
	const op = "services.event-sender.SendMessage"
	log := s.log.With(slog.String("op", op))

	log.Info("sending event", slog.Any("event", event))
}
