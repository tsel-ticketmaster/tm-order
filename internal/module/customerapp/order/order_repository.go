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

type OrderRepository interface {
	BeginTx(ctx context.Context) (*sql.Tx, error)
	CommitTx(ctx context.Context, tx *sql.Tx) error
	Rollback(ctx context.Context, tx *sql.Tx) error

	Save(ctx context.Context, o Order, tx *sql.Tx) error
	FindByID(ctx context.Context, ID string, tx *sql.Tx) (Order, error)
	FindMany(ctx context.Context, customerID int64, offset, limit int64, tx *sql.Tx) ([]Order, error)
	Count(ctx context.Context, customerID int64, tx *sql.Tx) (int64, error)
	Update(ctx context.Context, ID string, o Order, tx *sql.Tx) error
	CountActiveOrderByCustomerID(ctx context.Context, customerID int64, tx *sql.Tx) (int64, error)
}

type sqlCommand interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}

type orderRepository struct {
	logger *logrus.Logger
	db     *sql.DB
}

func NewOrderRepository(logger *logrus.Logger, db *sql.DB) OrderRepository {
	return &orderRepository{
		logger: logger,
		db:     db,
	}
}

// BeginTx implements OrderRepository.
func (r *orderRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return nil, errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred trying to begin transaction")
	}

	return tx, nil
}

// CommitTx implements OrderRepository.
func (r *orderRepository) CommitTx(ctx context.Context, tx *sql.Tx) error {
	if err := tx.Commit(); err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred trying to commit transaction")
	}

	return nil
}

// Rollback implements OrderRepository.
func (r *orderRepository) Rollback(ctx context.Context, tx *sql.Tx) error {
	if err := tx.Rollback(); err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred trying to rollback transaction")
	}

	return nil
}

// FindByID implements OrderRepository.
func (r *orderRepository) FindByID(ctx context.Context, ID string, tx *sql.Tx) (Order, error) {
	var cmd sqlCommand = r.db

	if tx != nil {
		cmd = tx
	}

	query := `
		SELECT 
			id, payment_method, transaction_id, virtual_account, status, customer_id, customer_name, customer_email,
			tax_percentage, service_charge_percentage, discount_percentage, service_charge,
			tax, discount, subtotal, total_amount, created_at, updated_at
		FROM ticket_order
		WHERE
			id = $1
		LIMIT 1
	`

	stmt, err := cmd.PrepareContext(ctx, query)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return Order{}, errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while getting order's prorperties")
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, ID)

	var data Order
	var virtualAccount sql.NullString
	var transactionID sql.NullString

	err = row.Scan(
		&data.ID, &data.PaymentMethod, &transactionID, &virtualAccount, &data.Status, &data.CustomerID, &data.CustomerName, &data.CustomerEmail,
		&data.TaxPercentage, &data.ServiceChargePercentage, &data.DiscountPercentage, &data.ServiceCharge,
		&data.Tax, &data.Discount, &data.Subtotal, &data.TotalAmount, &data.CreatedAt, &data.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return Order{}, errors.New(http.StatusNotFound, status.NOT_FOUND, fmt.Sprintf("order's properties with id '%s' is not found", ID))
		}
		r.logger.WithContext(ctx).WithError(err).Error()
		return Order{}, errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while getting order's prorperties")
	}

	if virtualAccount.Valid {
		data.VirtualAccount = &virtualAccount.String
	}
	if transactionID.Valid {
		data.TransactionID = &transactionID.String
	}

	return data, nil
}

// CountActiveOrderByCustomerID implements OrderRepository.
func (r *orderRepository) CountActiveOrderByCustomerID(ctx context.Context, customerID int64, tx *sql.Tx) (int64, error) {
	var cmd sqlCommand = r.db

	if tx != nil {
		cmd = tx
	}

	orderStatus := "WAITING_FOR_PAYMENT"

	query := `
		SELECT 
			count(id)
		FROM ticket_order
		WHERE
			customer_id = $1
		AND
			status = $2
		LIMIT 1
	`

	stmt, err := cmd.PrepareContext(ctx, query)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return 0, errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while counting order's prorperties")
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, customerID, orderStatus)

	var count int64

	err = row.Scan(&count)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return 0, errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while counting order's prorperties")
	}

	return count, nil
}

// FindMany implements OrderRepository.
func (r *orderRepository) FindMany(ctx context.Context, customerID int64, offset int64, limit int64, tx *sql.Tx) ([]Order, error) {
	var cmd sqlCommand = r.db

	if tx != nil {
		cmd = tx
	}

	query := `
		SELECT 
			id, payment_method, transaction_id, virtual_account, status, customer_id, customer_name, customer_email,
			tax_percentage, service_charge_percentage, discount_percentage, service_charge,
			tax, discount, subtotal, total_amount, created_at, updated_at
		FROM ticket_order
		WHERE
			customer_id = $1
		ORDER BY id DESC
		OFFSET $2
		LIMIT $3
	`

	stmt, err := cmd.PrepareContext(ctx, query)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return nil, errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while getting bunch of order's prorperties")
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, customerID, offset, limit)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return nil, errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while getting bunch of order's prorperties")
	}

	defer rows.Close()

	var data = make([]Order, 0)

	for rows.Next() {
		var o Order
		var virtualAccount sql.NullString
		var transactionID sql.NullString

		if err := rows.Scan(
			&o.ID, &o.PaymentMethod, &transactionID, &virtualAccount, &o.Status, &o.CustomerID, &o.CustomerName, &o.CustomerEmail,
			&o.TaxPercentage, &o.ServiceChargePercentage, &o.DiscountPercentage, &o.ServiceCharge,
			&o.Tax, &o.Discount, &o.Subtotal, &o.TotalAmount, &o.CreatedAt, &o.UpdatedAt,
		); err != nil {
			r.logger.WithContext(ctx).WithError(err).Error()
			return nil, errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while getting bunch of order's prorperties")
		}

		if virtualAccount.Valid {
			o.VirtualAccount = &virtualAccount.String
		}

		if transactionID.Valid {
			o.TransactionID = &transactionID.String
		}

		data = append(data, o)
	}

	return data, nil
}

// Count implements OrderRepository.
func (r *orderRepository) Count(ctx context.Context, customerID int64, tx *sql.Tx) (int64, error) {
	var cmd sqlCommand = r.db

	if tx != nil {
		cmd = tx
	}

	query := `
		SELECT count(id)
		FROM ticket_order
		WHERE
			customer_id = $1
		LIMIT 1
	`

	stmt, err := cmd.PrepareContext(ctx, query)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return 0, errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while getting order's prorperties")
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, customerID)

	var count int64

	err = row.Scan(
		&count,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, errors.New(http.StatusNotFound, status.NOT_FOUND, err.Error())
		}
		r.logger.WithContext(ctx).WithError(err).Error()
		return 0, errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while getting order's prorperties")
	}

	return count, nil
}

// Save implements OrderRepository.
func (r *orderRepository) Save(ctx context.Context, o Order, tx *sql.Tx) error {
	var cmd sqlCommand = r.db

	if tx != nil {
		cmd = tx
	}

	query := `
		INSERT INTO ticket_order
		(
			id, payment_method, status,
			customer_id, customer_name, customer_email, 
			tax_percentage, service_charge_percentage, discount_percentage,
			service_charge, tax, discount,
			subtotal, total_amount, created_at,
			updated_at, transaction_id, virtual_account
		)
		VALUES
		(
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
		)
	`

	stmt, err := cmd.PrepareContext(ctx, query)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while saving order's prorperties")
	}
	defer stmt.Close()

	var transactionID sql.NullString
	var virtualAccount sql.NullString

	if o.TransactionID != nil {
		transactionID.String = *o.TransactionID
		transactionID.Valid = true
	}

	if o.VirtualAccount != nil {
		virtualAccount.String = *o.VirtualAccount
		virtualAccount.Valid = true
	}

	_, err = stmt.ExecContext(ctx, o.ID, o.PaymentMethod, o.Status, o.CustomerID, o.CustomerName, o.CustomerEmail, o.TaxPercentage, o.ServiceChargePercentage,
		o.DiscountPercentage, o.ServiceCharge, o.Tax, o.Discount, o.Subtotal, o.TotalAmount, o.CreatedAt, o.UpdatedAt, transactionID, virtualAccount,
	)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while saving order's prorperties")
	}

	return nil
}

// Update implements OrderRepository.
func (r *orderRepository) Update(ctx context.Context, ID string, o Order, tx *sql.Tx) error {
	var cmd sqlCommand = r.db

	if tx != nil {
		cmd = tx
	}

	query := `
		UPDATE ticket_order
		SET
			status = $1,
			updated_at = $2
		WHERE id = $3
	`

	stmt, err := cmd.PrepareContext(ctx, query)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while updating order's prorperties")
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, o.Status, o.UpdatedAt, ID)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while updating order's prorperties")
	}

	return nil
}
