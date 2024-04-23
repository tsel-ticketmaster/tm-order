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
	BeginTx(ctx context.Context) (*sql.Tx, error)
	CommitTx(ctx context.Context, tx *sql.Tx) error
	Rollback(ctx context.Context, tx *sql.Tx) error

	Save(ctx context.Context, e Event, tx *sql.Tx) error
	FindByID(ctx context.Context, ID string, tx *sql.Tx) (Event, error)
	Update(ctx context.Context, ID string, update Event, tx *sql.DB) error
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

// BeginTx implements EventRepository.
func (r *eventRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return nil, errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred trying to begin transaction")
	}

	return tx, nil
}

// CommitTx implements EventRepository.
func (r *eventRepository) CommitTx(ctx context.Context, tx *sql.Tx) error {
	if err := tx.Commit(); err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred trying to commit transaction")
	}

	return nil
}

// Rollback implements EventRepository.
func (r *eventRepository) Rollback(ctx context.Context, tx *sql.Tx) error {
	if err := tx.Rollback(); err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred trying to rollback transaction")
	}

	return nil
}

// FindByID implements EventRepository.
func (r *eventRepository) FindByID(ctx context.Context, ID string, tx *sql.Tx) (Event, error) {
	var cmd sqlCommand = r.db

	if tx != nil {
		cmd = tx
	}

	query := `
		SELECT 
			id, name, description, status created_at, updated_at
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

// Save implements EventRepository.
func (r *eventRepository) Save(ctx context.Context, e Event, tx *sql.Tx) error {
	var cmd sqlCommand = r.db

	if tx != nil {
		cmd = tx
	}

	query := `
		INSERT INTO event 
		(
			id, name, description, status, created_at, updated_at
		)
		VALUES
		(
			$1, $2, $3, $4, $5, $6
		)
	`

	stmt, err := cmd.PrepareContext(ctx, query)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while saving event's prorperties")
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, e.ID, e.Name, e.Description, e.Status, e.CreatedAt, e.UpdatedAt)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while saving event's prorperties")
	}

	return nil
}

// Update implements EventRepository.
func (r *eventRepository) Update(ctx context.Context, ID string, e Event, tx *sql.DB) error {
	var cmd sqlCommand = r.db

	if tx != nil {
		cmd = tx
	}

	query := `
		UPDATE event
		SET
			name = $1,
			description = $2,
			status = $3,
			updated_at = $4
		WHERE id = $5
	`

	stmt, err := cmd.PrepareContext(ctx, query)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while updating event's prorperties")
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, e.Name, e.Description, e.Status, e.CreatedAt, e.UpdatedAt)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while updating event's prorperties")
	}

	return nil
}
