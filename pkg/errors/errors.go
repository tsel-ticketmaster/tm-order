package errors

import (
	"fmt"
	"net/http"

	"github.com/tsel-ticketmaster/tm-order/pkg/status"
)

type AppError struct {
	HTTPStatusCode int
	Status         string
	Message        string
}

// Error implements error.
func (e *AppError) Error() string {
	return fmt.Sprintf("%d %s: %s", e.HTTPStatusCode, e.Status, e.Message)
}

func New(httpStatusCode int, status string, message string) *AppError {
	if httpStatusCode < 400 && httpStatusCode > 599 {
		httpStatusCode = http.StatusInternalServerError
	}

	err := &AppError{
		HTTPStatusCode: httpStatusCode,
		Status:         status,
		Message:        message,
	}

	return err
}

func Destruct(err error) *AppError {
	if err == nil {
		return nil
	}
	ae, ok := err.(*AppError)
	if !ok {
		return New(http.StatusInternalServerError, status.INTERNAL_SERVER_ERROR, err.Error())
	}

	return ae
}

func MatchStatus(err error, s string) bool {
	if err == nil {
		return false
	}
	ae := Destruct(err)

	return ae.Status == s
}
