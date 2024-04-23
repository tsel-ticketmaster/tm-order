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

type ShowRepository interface {
	Save(ctx context.Context, s Show, tx *sql.Tx) error
	FindByID(ctx context.Context, ID string, tx *sql.Tx) (Show, error)
	FindManyByEventID(ctx context.Context, eventID string, tx *sql.Tx) ([]Show, error)
	Update(ctx context.Context, ID string, s Show, tx *sql.Tx) error
}

type showRepository struct {
	logger *logrus.Logger
	db     *sql.DB
}

func NewShowRepository(logger *logrus.Logger, db *sql.DB) ShowRepository {
	return &showRepository{
		logger: logger,
		db:     db,
	}
}

// FindByID implements ShowRepository.
func (r *showRepository) FindByID(ctx context.Context, ID string, tx *sql.Tx) (Show, error) {
	var cmd sqlCommand = r.db

	if tx != nil {
		cmd = tx
	}

	query := `
		SELECT 
			event_id, id, venue, type, time, status
		FROM event_show
		WHERE
			id = $1
		LIMIT 1
	`

	stmt, err := cmd.PrepareContext(ctx, query)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return Show{}, errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while getting event show's prorperties")
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, ID)

	var data Show
	err = row.Scan(
		&data.EventID, &data.ID, &data.Venue, &data.Type, &data.Time, &data.Status,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return Show{}, errors.New(http.StatusNotFound, status.NOT_FOUND, fmt.Sprintf("event show's properties with id '%s' is not found", ID))
		}
		r.logger.WithContext(ctx).WithError(err).Error()
		return Show{}, errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while getting event show's prorperties")
	}

	return data, nil
}

// FindManyByEventID implements ShowRepository.
func (r *showRepository) FindManyByEventID(ctx context.Context, eventID string, tx *sql.Tx) ([]Show, error) {
	var cmd sqlCommand = r.db

	if tx != nil {
		cmd = tx
	}

	query := `
		SELECT 
			event_id, id, venue, type, time, status
		FROM event_show
		WHERE
			event_id = $1
	`

	stmt, err := cmd.PrepareContext(ctx, query)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return nil, errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while getting bunch of event show's prorperties")
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, eventID)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return nil, errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while getting bunch of event show's prorperties")
	}

	defer rows.Close()

	var data = make([]Show, 0)
	for rows.Next() {
		var s Show

		err := rows.Scan(&s.EventID, &s.ID, &s.Venue, &s.Type, &s.Time, &s.Status)
		if err != nil {
			r.logger.WithContext(ctx).WithError(err).Error()
			return nil, errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while getting bunch of event show's prorperties")
		}

		data = append(data, s)
	}

	return data, nil
}

// Save implements ShowRepository.
func (r *showRepository) Save(ctx context.Context, s Show, tx *sql.Tx) error {
	var cmd sqlCommand = r.db

	if tx != nil {
		cmd = tx
	}

	query := `
		INSERT INTO event_show
		(
			event_id, id, venue, type, time, status
		)
		VALUES
		(
			$1, $2, $3, $4, $5, $6
		)
	`

	stmt, err := cmd.PrepareContext(ctx, query)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while saving event show's prorperties")
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, s.EventID, s.ID, s.Venue, s.Type, s.Time, s.Status)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while saving event show's prorperties")
	}

	return nil
}

// Update implements ShowRepository.
func (r *showRepository) Update(ctx context.Context, ID string, s Show, tx *sql.Tx) error {
	var cmd sqlCommand = r.db

	if tx != nil {
		cmd = tx
	}

	query := `
		UPDATE event_show
		SET
			venue = $1,
			type = $2,
			time = $3,
			status = $4
		WHERE id = $5
	`

	stmt, err := cmd.PrepareContext(ctx, query)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while updating event shows's prorperties")
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, s.Venue, s.Type, s.Time, s.Status)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while updating event show's prorperties")
	}

	return nil
}
