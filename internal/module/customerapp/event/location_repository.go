package event

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/tsel-ticketmaster/tm-order/pkg/errors"
	"github.com/tsel-ticketmaster/tm-order/pkg/status"
)

type LocationRepository interface {
	FindByShowID(ctx context.Context, showID string, tx *sql.Tx) (Location, error)
}

type locationRepository struct {
	logger *logrus.Logger
	db     *sql.DB
}

// FindByShowID implements LocationRepository.
func (r *locationRepository) FindByShowID(ctx context.Context, showID string, tx *sql.Tx) (Location, error) {
	var cmd sqlCommand = r.db

	if tx != nil {
		cmd = tx
	}

	query := `
		SELECT 
			event_id, show_id, country, city, formatted_address, latitude, longitude
		FROM event_show_location
		WHERE
			show_id = $1
		LIMIT 1
	`

	stmt, err := cmd.PrepareContext(ctx, query)
	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error()
		return Location{}, errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while getting event show location's prorperties")
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, showID)

	var data Location
	err = row.Scan(
		&data.EventID, &data.ShowID, &data.Country, &data.City, &data.FormattedAddress, &data.Latitude, &data.Longitude,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return Location{}, errors.New(http.StatusNotFound, status.NOT_FOUND, fmt.Sprintf("event show location's properties with id '%s' is not found", showID))
		}
		r.logger.WithContext(ctx).WithError(err).Error()
		return Location{}, errors.New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, "an error occurred while getting event show location's prorperties")
	}

	return data, nil
}

func NewLocationRepository(logger *logrus.Logger, db *sql.DB) LocationRepository {
	return &locationRepository{
		logger: logger,
		db:     db,
	}
}
