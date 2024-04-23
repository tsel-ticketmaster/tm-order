package event

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/tsel-ticketmaster/tm-order/pkg/errors"
	"github.com/tsel-ticketmaster/tm-order/pkg/status"
)

type ArtistRepository interface {
	FindManyByEventID(ctx context.Context, eventID string, tx *sql.Tx) ([]Artist, error)
	Save(ctx context.Context, a Artist, tx *sql.Tx) error
}

type artistRepository struct {
	logger *logrus.Logger
	db     *sql.DB
}

// FindByEventID implements ArtistRepository.
func (r *artistRepository) FindManyByEventID(ctx context.Context, eventID string, tx *sql.Tx) ([]Artist, error) {
	var cmd sqlCommand = r.db

	if tx != nil {
		cmd = tx
	}

	query := `
		SELECT 
			event_id, name
		FROM event_artist
		WHERE
			event_id = $1
	`

	stmt, err := cmd.PrepareContext(ctx, query)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return nil, errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while getting bunch of event artist's prorperties")
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, eventID)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return nil, errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while getting bunch of event artist's prorperties")
	}

	defer rows.Close()

	var data = make([]Artist, 0)
	for rows.Next() {
		var a Artist

		err := rows.Scan(&a.EventID, &a.Name)
		if err != nil {
			r.logger.WithContext(ctx).WithError(err).Error()
			return nil, errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while getting bunch of event artist's prorperties")
		}

		data = append(data, a)
	}

	return data, nil
}

// Save implements ArtistRepository.
func (r *artistRepository) Save(ctx context.Context, a Artist, tx *sql.Tx) error {
	var cmd sqlCommand = r.db

	if tx != nil {
		cmd = tx
	}

	query := `
		INSERT INTO event_artist
		(
			event_id, name
		)
		VALUES
		(
			$1, $2
		)
	`

	stmt, err := cmd.PrepareContext(ctx, query)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while saving event artist's prorperties")
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, a.EventID, a.Name)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while saving event artist's prorperties")
	}

	return nil
}

func NewArtistRepository(logger *logrus.Logger, db *sql.DB) ArtistRepository {
	return &artistRepository{
		logger: logger,
		db:     db,
	}
}
