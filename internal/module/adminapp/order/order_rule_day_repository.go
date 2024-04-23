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
	Save(ctx context.Context, rule OrderRuleDay, tx *sql.Tx) error
	Delete(ctx context.Context, eventID string, tx *sql.Tx) error
}

type orderRuleDayRepository struct {
	logger *logrus.Logger
	db     *sql.DB
}

// Delete implements OrderRuleDayRepository.
func (r *orderRuleDayRepository) Delete(ctx context.Context, eventID string, tx *sql.Tx) error {
	var cmd sqlCommand = r.db

	if tx != nil {
		cmd = tx
	}

	query := `
		DELETE FROM order_rule_day WHERE event_id = $1
	`

	stmt, err := cmd.PrepareContext(ctx, query)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while deleting order rule day's prorperties")
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, eventID)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while deleting order rule days's prorperties")
	}

	return nil
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

// Save implements OrderRuleDayRepository.
func (r *orderRuleDayRepository) Save(ctx context.Context, rule OrderRuleDay, tx *sql.Tx) error {
	var cmd sqlCommand = r.db

	if tx != nil {
		cmd = tx
	}

	query := `
		INSERT INTO order_rule_day
		(
			event_id, day
		)
		VALUES
		(
			$1, $2
		)
	`

	stmt, err := cmd.PrepareContext(ctx, query)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while saving order rule day's prorperties")
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, rule.EventID, rule.Day)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while saving order rule days's prorperties")
	}

	return nil
}

func NewOrderRuleDayRepository(logger *logrus.Logger, db *sql.DB) OrderRuleDayRepository {
	return &orderRuleDayRepository{
		logger: logger,
		db:     db,
	}
}
