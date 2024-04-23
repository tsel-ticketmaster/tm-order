package event

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/tsel-ticketmaster/tm-order/internal/pkg/middleware"
	"github.com/tsel-ticketmaster/tm-order/pkg/errors"
	publicMiddleware "github.com/tsel-ticketmaster/tm-order/pkg/middleware"
	"github.com/tsel-ticketmaster/tm-order/pkg/response"
	"github.com/tsel-ticketmaster/tm-order/pkg/status"
)

type HTTPHandler struct {
	SessionMiddleware *middleware.AdminSession
	Validate          *validator.Validate
	EventUseCase      EventUseCase
}

func InitHTTPHandler(router *mux.Router, adminSession *middleware.AdminSession, validate *validator.Validate, eventUsecase EventUseCase) {
	handler := &HTTPHandler{
		Validate:     validate,
		EventUseCase: eventUsecase,
	}

	router.HandleFunc("/tm-order/v1/adminapp/events", publicMiddleware.SetRouteChain(handler.CreateEvent, adminSession.Verify)).Methods(http.MethodPost)
}

func (handler HTTPHandler) validate(ctx context.Context, payload interface{}) error {
	err := handler.Validate.StructCtx(ctx, payload)
	if err == nil {
		return nil
	}

	errorFields := err.(validator.ValidationErrors)

	errMessages := make([]string, len(errorFields))

	for k, errorField := range errorFields {
		errMessages[k] = fmt.Sprintf("invalid '%s' with value '%v'", errorField.Field(), errorField.Value())
	}

	errorMessage := strings.Join(errMessages, ", ")

	return fmt.Errorf(errorMessage)

}

func (handler HTTPHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req := CreateEventRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSON(w, http.StatusUnprocessableEntity, response.RESTEnvelope{
			Status:  status.UNPROCESSABLE_ENTITY,
			Message: err.Error(),
		})

		return
	}

	if err := handler.validate(ctx, req); err != nil {
		response.JSON(w, http.StatusBadRequest, response.RESTEnvelope{
			Status:  status.BAD_REQUEST,
			Message: err.Error(),
		})

		return
	}

	resp, err := handler.EventUseCase.CreateEvent(ctx, req)
	if err != nil {
		ae := errors.Destruct(err)
		response.JSON(w, ae.HTTPStatusCode, response.RESTEnvelope{
			Status:  ae.Status,
			Message: ae.Message,
		})

		return
	}
	response.JSON(w, http.StatusCreated, response.RESTEnvelope{
		Status:  status.CREATED,
		Message: "event has been successfully created",
		Data:    resp,
		Meta:    nil,
	})

}
