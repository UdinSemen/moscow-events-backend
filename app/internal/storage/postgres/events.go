package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/UdinSemen/moscow-events-backend/internal/domain/models"
	"golang.org/x/net/context"
)

const opPrefixPgStorageEvents = "pg_storage.events."

var (
	ErrInvalidDates = errors.New("invalid dates")
)

func (s *PgStorage) GetEvents(ctx context.Context, userID, category string, date []time.Time) ([]models.Event, error) {
	const op = opPrefixPgStorageEvents + "GetEvents"

	if !slices.Contains([]int{1, 2}, len(date)) {
		return nil, fmt.Errorf("%s:%w", op, ErrInvalidDates)
	}
	stmt := "where d.date = :date_fir"
	args := map[string]interface{}{
		"date_fir": date[0],
		"cat":      category,
		"user_id":  userID,
	}
	if len(date) == 2 {
		stmt = "where d.date between :date_fir and :date_sec"
		args["date_sec"] = date[1]
	}

	query := fmt.Sprint("SELECT ev.id, label, description, d.date, ev.price, coalesce(ev.url_buy, '') AS url_buy, url_img, CASE WHEN fv.id_event IS NOT NULL THEN TRUE ELSE FALSE END AS is_favorite" +
		" FROM public.news_events ev JOIN public.dates d ON ev.id = d.id_event" +
		" LEFT JOIN favourite_list fv ON ev.id = fv.id_event and fv.id_date = d.id and fv.user_id =:user_id " +
		stmt +
		" AND url_img NOTNULL AND price NOTNULL AND label NOTNULL and ev.url_img NOTNULL and ev.description NOTNULL AND ev.category =:cat" +
		" and id_group in (select id_group from public.news_events_actual_group)" +
		" ORDER BY date")

	rows, err := s.db.NamedQueryContext(ctx, query, args)
	if err != nil {
		outErr := fmt.Errorf("%s:%w", op, err)
		if errors.Is(err, sql.ErrNoRows) {
			outErr = fmt.Errorf("%s:%w", op, ErrNoRows)
		}
		return nil, outErr
	}

	var events []models.Event

	for rows.Next() {
		var event models.Event
		if err := rows.StructScan(&event); err != nil {
			return nil, fmt.Errorf("%s:%w", op, err)
		}
		events = append(events, event)
	}

	return events, nil
}
