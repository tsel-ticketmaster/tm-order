package order

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/tsel-ticketmaster/tm-order/pkg/errors"
	"github.com/tsel-ticketmaster/tm-order/pkg/status"
)

type OrderRuleDayRepository interface {
	FindManyByEventID(ctx context.Context, eventID string, tx *sql.Tx) ([]OrderRuleDay, error)
}

type orderRuleDayRepository struct {
	logger *logrus.Logger
	db     *sql.DB
}

// FindManyByEventID implements OrderRuleDayRepository.
func (r *orderRuleDayRepository) FindManyByEventID(ctx context.Context, eventID string, tx *sql.Tx) ([]OrderRuleDay, error) {
	var cmd sqlCommand = r.db

	if tx != nil {
		cmd = tx
	}

	query := `
		SELECT 
			event_id, day
		FROM order_rule_day
		WHERE
			event_id = $1
	`

	stmt, err := cmd.PrepareContext(ctx, query)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return nil, errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while getting bunch of order rule day's prorperties")
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, eventID)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return nil, errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while getting bunch of order rule day's prorperties")
	}

	defer rows.Close()

	var data = make([]OrderRuleDay, 0)
	for rows.Next() {
		var rule OrderRuleDay

		err := rows.Scan(&rule.EventID, &rule.Day)
		if err != nil {
			r.logger.WithContext(ctx).WithError(err).Error()
			return nil, errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while getting bunch of order rule day's prorperties")
		}

		data = append(data, rule)
	}

	return data, nil
}

func NewOrderRuleDayRepository(logger *logrus.Logger, db *sql.DB) OrderRuleDayRepository {
	return &orderRuleDayRepository{
		logger: logger,
		db:     db,
	}
}
