package order

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/tsel-ticketmaster/tm-order/pkg/errors"
	"github.com/tsel-ticketmaster/tm-order/pkg/status"
)

type OrderRuleRangeDateRepository interface {
	FindByEventID(ctx context.Context, eventID string, tx *sql.Tx) (OrderRuleRangeDate, error)
}

type orderRuleRangeDateRepository struct {
	logger *logrus.Logger
	db     *sql.DB
}

func NewOrderRuleRangeDateRepository(logger *logrus.Logger, db *sql.DB) OrderRuleRangeDateRepository {
	return &orderRuleRangeDateRepository{
		logger: logger,
		db:     db,
	}
}

// FindByEventID implements OrderRuleRangeDateRepository.
func (r *orderRuleRangeDateRepository) FindByEventID(ctx context.Context, eventID string, tx *sql.Tx) (OrderRuleRangeDate, error) {
	var cmd sqlCommand = r.db

	if tx != nil {
		cmd = tx
	}

	query := `
		SELECT 
			event_id, start_date, end_date
		FROM order_rule_range_date
		WHERE
			event_id = $1
		LIMIT 1
	`

	stmt, err := cmd.PrepareContext(ctx, query)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return OrderRuleRangeDate{}, errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while getting order rule range date's prorperties")
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, eventID)

	var data OrderRuleRangeDate
	err = row.Scan(
		&data.EventID, &data.StartDate, &data.EndDate,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return OrderRuleRangeDate{}, errors.New(http.StatusNotFound, status.NOT_FOUND, fmt.Sprintf("order rule range date's properties with id '%s' is not found", eventID))
		}
		r.logger.WithContext(ctx).WithError(err).Error()
		return OrderRuleRangeDate{}, errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while getting order rule range date's prorperties")
	}

	return data, nil
}
