package event

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/tsel-ticketmaster/tm-order/pkg/errors"
	"github.com/tsel-ticketmaster/tm-order/pkg/status"
)

type PromotorRepository interface {
	FindManyByEventID(ctx context.Context, eventID string, tx *sql.Tx) ([]Promotor, error)
	Save(ctx context.Context, p Promotor, tx *sql.Tx) error
}

type promotorRepository struct {
	logger *logrus.Logger
	db     *sql.DB
}

func NewPromotorRepository(logger *logrus.Logger, db *sql.DB) PromotorRepository {
	return &promotorRepository{
		logger: logger,
		db:     db,
	}
}

// FindManyByEventID implements PromotorRepository.
func (r *promotorRepository) FindManyByEventID(ctx context.Context, eventID string, tx *sql.Tx) ([]Promotor, error) {
	var cmd sqlCommand = r.db

	if tx != nil {
		cmd = tx
	}

	query := `
		SELECT 
			event_id, name, email, phone
		FROM event_promotor
		WHERE
			event_id = $1
	`

	stmt, err := cmd.PrepareContext(ctx, query)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return nil, errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while getting bunch of event promotor's prorperties")
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, eventID)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return nil, errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while getting bunch of event promotor's prorperties")
	}

	defer rows.Close()

	var data = make([]Promotor, 0)
	for rows.Next() {
		var p Promotor

		err := rows.Scan(&p.EventID, &p.Name, &p.Email, &p.Phone)
		if err != nil {
			r.logger.WithContext(ctx).WithError(err).Error()
			return nil, errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while getting bunch of event promotor's prorperties")
		}

		data = append(data, p)
	}

	return data, nil
}

// Save implements PromotorRepository.
func (r *promotorRepository) Save(ctx context.Context, p Promotor, tx *sql.Tx) error {
	var cmd sqlCommand = r.db

	if tx != nil {
		cmd = tx
	}

	query := `
		INSERT INTO event_promotor
		(
			event_id, name, email, phone
		)
		VALUES
		(
			$1, $2, $3, $4
		)
	`

	stmt, err := cmd.PrepareContext(ctx, query)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while saving event artist's prorperties")
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, p.EventID, p.Name, p.Email, p.Phone)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while saving event artist's prorperties")
	}

	return nil
}
