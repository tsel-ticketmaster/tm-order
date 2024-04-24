package ticket

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/tsel-ticketmaster/tm-order/pkg/errors"
	"github.com/tsel-ticketmaster/tm-order/pkg/status"
)

type AcquiredTicketRepository interface {
	CountByEventIDAndCustomerID(ctx context.Context, eventID string, customerID int64, tx *sql.Tx) (int64, error)
}

type acquiredTicketRepository struct {
	logger *logrus.Logger
	db     *sql.DB
}

func NewAcquiredTicketRepository(logger *logrus.Logger, db *sql.DB) AcquiredTicketRepository {
	return &acquiredTicketRepository{
		logger: logger,
		db:     db,
	}
}

// CountByEventIDAndCustomerID implements AcquiredTicketRepository.
func (r *acquiredTicketRepository) CountByEventIDAndCustomerID(ctx context.Context, eventID string, customerID int64, tx *sql.Tx) (int64, error) {
	var cmd sqlCommand = r.db

	if tx != nil {
		cmd = tx
	}

	query := `SELECT count(id) FROM acquired_ticket WHERE event_id = $1 AND customer_id = $2`

	stmt, err := cmd.PrepareContext(ctx, query)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return 0, errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while counting acquired ticket's prorperties")
	}
	defer stmt.Close()

	var count int64
	row := stmt.QueryRowContext(ctx, eventID, customerID)

	err = row.Scan(&count)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return 0, errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while counting acquired ticket's prorperties for update")
	}
	return count, nil
}
