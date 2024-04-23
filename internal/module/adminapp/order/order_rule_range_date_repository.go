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
	Save(ctx context.Context, rule OrderRuleRangeDate, tx *sql.Tx) error
	Update(ctx context.Context, eventID string, rule OrderRuleRangeDate, tx *sql.Tx) error
}

type sqlCommand interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
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

// Save implements OrderRuleRangeDateRepository.
func (r *orderRuleRangeDateRepository) Save(ctx context.Context, rule OrderRuleRangeDate, tx *sql.Tx) error {
	var cmd sqlCommand = r.db

	if tx != nil {
		cmd = tx
	}

	query := `
		INSERT INTO order_rule_range_date
		(
			event_id, start_date, end_date
		)
		VALUES
		(
			$1, $2, $3
		)
	`

	stmt, err := cmd.PrepareContext(ctx, query)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while saving order rule range date's prorperties")
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, rule.EventID, rule.StartDate, rule.EndDate)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while saving order rule range date's prorperties")
	}

	return nil
}

// Update implements OrderRuleRangeDateRepository.
func (r *orderRuleRangeDateRepository) Update(ctx context.Context, eventID string, rule OrderRuleRangeDate, tx *sql.Tx) error {
	var cmd sqlCommand = r.db

	if tx != nil {
		cmd = tx
	}

	query := `
		UPDATE order_rule_range_date
		SET
			start_date = $2,
			end_date = $3,
		WHERE event_id = $4
	`

	stmt, err := cmd.PrepareContext(ctx, query)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while updating order rule range date's prorperties")
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, rule.StartDate, rule.EndDate)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while updating order rule range date's prorperties")
	}

	return nil
}
