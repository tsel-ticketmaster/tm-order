package ticket

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/tsel-ticketmaster/tm-order/pkg/errors"
	"github.com/tsel-ticketmaster/tm-order/pkg/status"
)

type TicketStockRepository interface {
	FindManyByShowID(ctx context.Context, showID string, tx *sql.Tx) ([]TicketStock, error)
	FindByIDForUpdate(ctx context.Context, ID string, tx *sql.Tx) (TicketStock, error)
	Update(ctx context.Context, ID string, ts TicketStock, tx *sql.Tx) error
}

type sqlCommand interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}

type ticketStockRepository struct {
	logger *logrus.Logger
	db     *sql.DB
}

func NewTicketStockRepository(logger *logrus.Logger, db *sql.DB) TicketStockRepository {
	return &ticketStockRepository{
		logger: logger,
		db:     db,
	}
}

// FindByIDForUpdate implements TicketStockRepository.
func (r *ticketStockRepository) FindByIDForUpdate(ctx context.Context, ID string, tx *sql.Tx) (TicketStock, error) {
	var cmd sqlCommand = r.db

	if tx != nil {
		cmd = tx
	}

	query := `
		SELECT 
			id, tier, allocation, price, acquired, last_stock_update, online_for, show_id, event_id
		FROM ticket_stock
		WHERE
			id = $1
		FOR UPDATE
	`

	stmt, err := cmd.PrepareContext(ctx, query)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return TicketStock{}, errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while getting ticket stock's prorperties for update")
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, ID)

	var data TicketStock
	var onlineFor sql.NullString

	err = row.Scan(&data.ID, &data.Tier, &data.Allocation, &data.Price, &data.Acquired, &data.LastStockUpdate, &onlineFor, &data.ShowID, &data.EventID)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return TicketStock{}, errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while getting ticket stock's prorperties for update")
	}
	return data, nil
}

// Update implements TicketStockRepository.
func (r *ticketStockRepository) Update(ctx context.Context, ID string, ts TicketStock, tx *sql.Tx) error {
	var cmd sqlCommand = r.db

	if tx != nil {
		cmd = tx
	}

	query := `
		UPDATE ticket_stock
		SET
			acquired = $1,
			last_stock_update = $2
		WHERE 
			id = $3
	`

	stmt, err := cmd.PrepareContext(ctx, query)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while updating ticket stock's prorperties")
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, ts.Acquired, ts.LastStockUpdate, ID)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while updating ticket stock's prorperties")
	}

	return nil
}

// FindManyByShowID implements TicketStockRepository.
func (r *ticketStockRepository) FindManyByShowID(ctx context.Context, showID string, tx *sql.Tx) ([]TicketStock, error) {
	var cmd sqlCommand = r.db

	if tx != nil {
		cmd = tx
	}

	query := `
		SELECT 
			id, tier, allocation, price, acquired, last_stock_update, online_for, show_id, event_id
		FROM ticket_stock
		WHERE
			show_id = $1
	`

	stmt, err := cmd.PrepareContext(ctx, query)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return nil, errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while getting bunch of ticket stock's prorperties")
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, showID)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return nil, errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while getting bunch of ticket stock's prorperties")
	}

	defer rows.Close()

	var data = make([]TicketStock, 0)
	for rows.Next() {
		var ts TicketStock
		var onlineFor sql.NullString
		err := rows.Scan(&ts.ID, &ts.Tier, &ts.Allocation, &ts.Price, &ts.Acquired, &ts.LastStockUpdate, &onlineFor, &ts.ShowID, &ts.EventID)
		if err != nil {
			r.logger.WithContext(ctx).WithError(err).Error()
			return nil, errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while getting bunch of order rule day's prorperties")
		}

		if onlineFor.Valid {
			ts.OnlineFor = &onlineFor.String
		}

		data = append(data, ts)
	}

	return data, nil
}

// Save implements TicketStockRepository.
func (r *ticketStockRepository) Save(ctx context.Context, ts TicketStock, tx *sql.Tx) error {
	var cmd sqlCommand = r.db

	if tx != nil {
		cmd = tx
	}

	query := `
		INSERT INTO ticket_stock
		(
			id, tier, allocation, price, acquired, last_stock_update, online_for, show_id, event_id
		)
		VALUES
		(
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)
	`

	stmt, err := cmd.PrepareContext(ctx, query)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while saving ticket stock's prorperties")
	}
	defer stmt.Close()

	var onlineFor sql.NullString
	if ts.OnlineFor != nil {
		onlineFor.Valid = true
		onlineFor.String = *ts.OnlineFor
	}

	_, err = stmt.ExecContext(ctx, ts.ID, ts.Tier, ts.Allocation, ts.Price, ts.Acquired, ts.LastStockUpdate, onlineFor, ts.ShowID, ts.EventID)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while saving ticket stock's prorperties")
	}

	return nil
}
