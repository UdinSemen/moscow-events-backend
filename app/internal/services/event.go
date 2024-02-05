package services

import (
	"fmt"
	"time"

	"github.com/UdinSemen/moscow-events-backend/internal/domain/models"
	"github.com/UdinSemen/moscow-events-backend/internal/storage"
	"golang.org/x/net/context"
)

const eventServiceOpPrefix = "services.event."

type EventService struct {
	postgres storage.PgStorage
}

func NewEventService(postgres storage.PgStorage) *EventService {
	return &EventService{postgres: postgres}
}

func (s *EventService) GetEvents(ctx context.Context, userID, category string, date []time.Time) ([]models.Event, error) {
	const op = opAuthServPrefix + "GetEvents"

	events, err := s.postgres.GetEvents(ctx, userID, category, date)
	if err != nil {
		return nil, fmt.Errorf("%s:%w", op, err)
	}
	return events, nil
}
