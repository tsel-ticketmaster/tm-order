package event

import (
	"context"
	"database/sql"

	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/tsel-ticketmaster/tm-order/pkg/errors"
	"github.com/tsel-ticketmaster/tm-order/pkg/status"
)

type EventRepository interface {
	FindByID(ctx context.Context, ID string, tx *sql.Tx) (Event, error)
}

type sqlCommand interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}

type eventRepository struct {
	logger *logrus.Logger
	db     *sql.DB
}

func NewEventRepository(logger *logrus.Logger, db *sql.DB) EventRepository {
	return &eventRepository{
		logger: logger,
		db:     db,
	}
}

// FindByID implements EventRepository.
func (r *eventRepository) FindByID(ctx context.Context, ID string, tx *sql.Tx) (Event, error) {
	var cmd sqlCommand = r.db

	if tx != nil {
		cmd = tx
	}

	query := `
		SELECT 
			id, name, description, status, created_at, updated_at
		FROM event
		WHERE
			id = $1
		LIMIT 1
	`

	stmt, err := cmd.PrepareContext(ctx, query)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return Event{}, errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while getting event's prorperties")
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, ID)

	var data Event
	err = row.Scan(
		&data.ID, &data.Name, &data.Description, &data.Status, &data.CreatedAt, &data.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return Event{}, errors.New(http.StatusNotFound, status.NOT_FOUND, fmt.Sprintf("event's properties with id '%s' is not found", ID))
		}
		r.logger.WithContext(ctx).WithError(err).Error()
		return Event{}, errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while getting event's prorperties")
	}

	return data, nil
}
