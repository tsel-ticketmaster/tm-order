package order

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/tsel-ticketmaster/tm-order/pkg/errors"
	"github.com/tsel-ticketmaster/tm-order/pkg/status"
)

type ItemRepository interface {
	FindManyByOrderID(ctx context.Context, orderID string, tx *sql.Tx) ([]Item, error)
	Save(ctx context.Context, i Item, tx *sql.Tx) error
}

type itemRepository struct {
	logger *logrus.Logger
	db     *sql.DB
}

func NewItemRepository(logger *logrus.Logger, db *sql.DB) ItemRepository {
	return &itemRepository{
		logger: logger,
		db:     db,
	}
}

// FindManyByOrderID implements ItemRepository.
func (r *itemRepository) FindManyByOrderID(ctx context.Context, orderID string, tx *sql.Tx) ([]Item, error) {
	var cmd sqlCommand = r.db

	if tx != nil {
		cmd = tx
	}

	query := `
		SELECT 
			id, order_id, ticket_stock_id, tier, show_id, show_venue, event_id, event_name, price, quantity
		FROM order_item
		WHERE
			order_id = $1
	`

	stmt, err := cmd.PrepareContext(ctx, query)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return nil, errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while getting bunch of order item's prorperties")
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, orderID)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return nil, errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while getting bunch of order item's prorperties")
	}

	defer rows.Close()

	var data = make([]Item, 0)

	for rows.Next() {
		var i Item

		if err := rows.Scan(
			&i.ID, &i.OrderID, &i.TicketStockID, &i.Tier, &i.ShowID, &i.ShowVenue, &i.EventID, &i.EventName, &i.Price, &i.Quantity,
		); err != nil {
			r.logger.WithContext(ctx).WithError(err).Error()
			return nil, errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while getting bunch of order item's prorperties")
		}

		data = append(data, i)
	}

	return data, nil
}

// Save implements ItemRepository.
func (r *itemRepository) Save(ctx context.Context, i Item, tx *sql.Tx) error {
	var cmd sqlCommand = r.db

	if tx != nil {
		cmd = tx
	}

	query := `
		INSERT INTO order_item
		(
			order_id, ticket_stock_id, show_id, event_id, event_name, show_venue, tier, price, quantity
		)
		VALUES
		(
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)
	`

	stmt, err := cmd.PrepareContext(ctx, query)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while saving order item's prorperties")
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, i.OrderID, i.TicketStockID, i.ShowID, i.EventID, i.EventName, i.ShowVenue, i.Tier, i.Price, i.Quantity)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while saving order items's prorperties")
	}

	return nil
}
