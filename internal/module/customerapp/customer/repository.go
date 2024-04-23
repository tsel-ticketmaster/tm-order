package customer

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/tsel-ticketmaster/tm-order/pkg/errors"
	"github.com/tsel-ticketmaster/tm-order/pkg/status"
)

type CustomerRepository interface {
	Save(ctx context.Context, c Customer, tx *sql.Tx) (int64, error)
	FindByID(ctx context.Context, ID int64, tx *sql.Tx) (Customer, error)
	FindByEmail(ctx context.Context, email string, tx *sql.Tx) (Customer, error)
	Update(ctx context.Context, ID int64, update Customer, tx *sql.Tx) error
}

type sqlCommand interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}

type customerRepository struct {
	logger *logrus.Logger
	db     *sql.DB
}

// FindByEmail implements CustomerRepository.
func (r *customerRepository) FindByEmail(ctx context.Context, email string, tx *sql.Tx) (Customer, error) {
	var cmd sqlCommand = r.db

	if tx != nil {
		cmd = tx
	}

	query := `
		SELECT 
			id, name, email, password, password_salt, verification_status, member_status, created_at, updated_at
		FROM customer
		WHERE
			email = $1
		LIMIT 1
	`

	stmt, err := cmd.PrepareContext(ctx, query)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return Customer{}, errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while getting customer's prorperties")
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, email)

	var data Customer

	err = row.Scan(
		&data.ID, &data.Name, &data.Email, &data.Password, &data.PasswordSalt, &data.VerificationStatus, &data.MemberStatus, &data.CreatedAt, &data.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return Customer{}, errors.New(http.StatusNotFound, status.NOT_FOUND, fmt.Sprintf("customer's properties with email '%s' is not found", email))
		}
		r.logger.WithContext(ctx).WithError(err).Error()
		return Customer{}, errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while getting customer's prorperties")
	}

	return data, nil
}

// FindByID implements CustomerRepository.
func (r *customerRepository) FindByID(ctx context.Context, ID int64, tx *sql.Tx) (Customer, error) {
	var cmd sqlCommand = r.db

	if tx != nil {
		cmd = tx
	}

	query := `
		SELECT 
			id, name, email, password, password_salt, verification_status, member_status, created_at, updated_at
		FROM customer
		WHERE
			id = $1
		LIMIT 1
	`

	stmt, err := cmd.PrepareContext(ctx, query)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return Customer{}, errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while getting customer's prorperties")
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, ID)

	var data Customer

	err = row.Scan(
		&data.ID, &data.Name, &data.Email, &data.Password, &data.PasswordSalt, &data.VerificationStatus, &data.MemberStatus, &data.CreatedAt, &data.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return Customer{}, errors.New(http.StatusNotFound, status.NOT_FOUND, fmt.Sprintf("customer's properties with id '%d' is not found", ID))
		}
		r.logger.WithContext(ctx).WithError(err).Error()
		return Customer{}, errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while getting customer's prorperties")
	}

	return data, nil
}

// Save implements CustomerRepository.
func (r *customerRepository) Save(ctx context.Context, c Customer, tx *sql.Tx) (int64, error) {
	var cmd sqlCommand = r.db

	if tx != nil {
		cmd = tx
	}

	query := `
		INSERT INTO customer
		(
			name, email, password, password_salt, verification_status, member_status, created_at, updated_at
		)
		VALUES
		(
			$1, $2, $3, $4, $5, $6, $7, $8
		)
		RETURNING id
	`

	stmt, err := cmd.PrepareContext(ctx, query)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return 0, errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while saving customer's prorperties")
	}

	row := stmt.QueryRowContext(ctx, c.Name, c.Email, c.Password, c.PasswordSalt, c.VerificationStatus, c.MemberStatus, c.CreatedAt, c.UpdatedAt)

	var ID int64

	err = row.Scan(&ID)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return 0, errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while saving customer's prorperties")
	}

	return ID, nil
}

// Update implements CustomerRepository.
func (r *customerRepository) Update(ctx context.Context, ID int64, c Customer, tx *sql.Tx) error {
	var cmd sqlCommand = r.db

	if tx != nil {
		cmd = tx
	}

	query := `
		UPDATE customer
		SET
			name = $1,
			email = $2,
			password = $3,
			password_salt = $4,
			verification_status = $5,
			member_status = $6,
			updated_at = $7
		WHERE
			id = $8
	`

	stmt, err := cmd.PrepareContext(ctx, query)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while updating customer's prorperties")
	}

	_, err = stmt.ExecContext(ctx, c.Name, c.Email, c.Password, c.PasswordSalt, c.VerificationStatus, c.MemberStatus, c.UpdatedAt, ID)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while updating customer's prorperties")
	}

	return nil
}

func NewCustomerRepository(logger *logrus.Logger, db *sql.DB) CustomerRepository {
	return &customerRepository{
		logger: logger,
		db:     db,
	}
}
