package order

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
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
	SessionMiddleware *middleware.CustomerSession
	Validate          *validator.Validate
	OrderUseCase      OrderUseCase
}

func InitHTTPHandler(router *mux.Router, customerSession *middleware.CustomerSession, validate *validator.Validate, orderUseCase OrderUseCase) {
	handler := &HTTPHandler{
		Validate:     validate,
		OrderUseCase: orderUseCase,
	}

	router.HandleFunc("/tm-order/v1/customerapp/orders", publicMiddleware.SetRouteChain(handler.PlaceOrder, customerSession.Verify)).Methods(http.MethodPost)
	router.HandleFunc("/tm-order/v1/customerapp/orders", publicMiddleware.SetRouteChain(handler.GetManyOrder, customerSession.Verify)).Methods(http.MethodGet)
	router.HandleFunc("/tm-order/v1/customerapp/orders/on-expire", publicMiddleware.SetRouteChain(handler.OnExpireOrder)).Methods(http.MethodPost)
	router.HandleFunc("/tm-order/v1/customerapp/orders/on-payment-notification", publicMiddleware.SetRouteChain(handler.OnPaymentNotification)).Methods(http.MethodPost)
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

func (handler HTTPHandler) GetManyOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	qs := r.URL.Query()

	req := GetManyOrderRequest{}
	req.Page, _ = strconv.ParseInt(qs.Get("page"), 10, 64)
	req.Size, _ = strconv.ParseInt(qs.Get("size"), 10, 64)

	if err := handler.validate(ctx, req); err != nil {
		response.JSON(w, http.StatusBadRequest, response.RESTEnvelope{
			Status:  status.BAD_REQUEST,
			Message: err.Error(),
		})

		return
	}

	resp, err := handler.OrderUseCase.GetManyOrder(ctx, req)
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
		Message: "list of orders",
		Data:    resp,
		Meta:    nil,
	})

}

func (handler HTTPHandler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req := PlaceOrderRequest{}
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

	resp, err := handler.OrderUseCase.PlaceOrder(ctx, req)
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
		Message: "order has been successfully placed",
		Data:    resp,
		Meta:    nil,
	})

}

func (handler HTTPHandler) OnExpireOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	e := ExpireOrderEvent{}
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		response.JSON(w, http.StatusUnprocessableEntity, response.RESTEnvelope{
			Status:  status.UNPROCESSABLE_ENTITY,
			Message: err.Error(),
		})

		return
	}

	err := handler.OrderUseCase.OnExpireOrder(ctx, e)
	if err != nil {
		ae := errors.Destruct(err)
		response.JSON(w, ae.HTTPStatusCode, response.RESTEnvelope{
			Status:  ae.Status,
			Message: ae.Message,
		})

		return
	}
	response.JSON(w, http.StatusOK, response.RESTEnvelope{
		Status:  status.OK,
		Message: "order has been successfully expired",
		Data:    nil,
		Meta:    nil,
	})

}

func (handler HTTPHandler) OnPaymentNotification(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	e := PaymentNotificationEvent{}
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		response.JSON(w, http.StatusUnprocessableEntity, response.RESTEnvelope{
			Status:  status.UNPROCESSABLE_ENTITY,
			Message: err.Error(),
		})

		return
	}

	err := handler.OrderUseCase.OnPaymentNotification(ctx, e)
	if err != nil {
		ae := errors.Destruct(err)
		response.JSON(w, ae.HTTPStatusCode, response.RESTEnvelope{
			Status:  ae.Status,
			Message: ae.Message,
		})

		return
	}
	response.JSON(w, http.StatusOK, response.RESTEnvelope{
		Status:  status.OK,
		Message: "order has been update by payment notification",
		Data:    nil,
		Meta:    nil,
	})

}
