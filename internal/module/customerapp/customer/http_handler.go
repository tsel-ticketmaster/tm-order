package customer

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
	CustomerUseCase   CustomerUseCase
}

func InitHTTPHandler(router *mux.Router, customerSession *middleware.CustomerSession, validate *validator.Validate, customerUseCase CustomerUseCase) {
	handler := &HTTPHandler{
		Validate:        validate,
		CustomerUseCase: customerUseCase,
	}

	router.HandleFunc("/tm-order/v1/customerapp/customers/signin", publicMiddleware.SetRouteChain(handler.SignIn)).Methods(http.MethodPost)
	router.HandleFunc("/tm-order/v1/customerapp/customers/signup", publicMiddleware.SetRouteChain(handler.SignUp)).Methods(http.MethodPost)
	router.HandleFunc("/tm-order/v1/customerapp/customers/signout", publicMiddleware.SetRouteChain(handler.SignOut, customerSession.Verify)).Methods(http.MethodPost)
	router.HandleFunc("/tm-order/v1/customerapp/customers/profile", publicMiddleware.SetRouteChain(handler.GetProfile, customerSession.Verify)).Methods(http.MethodGet)
	router.HandleFunc("/tm-order/v1/customerapp/customers/profile", publicMiddleware.SetRouteChain(handler.UpdateProfile, customerSession.Verify)).Methods(http.MethodPatch)
	router.HandleFunc("/tm-order/v1/customerapp/customers/change-email", publicMiddleware.SetRouteChain(handler.ChangeEmail, customerSession.Verify)).Methods(http.MethodPatch)
	router.HandleFunc("/tm-order/v1/customerapp/customers/change-password", publicMiddleware.SetRouteChain(handler.ChangePassword, customerSession.Verify)).Methods(http.MethodPatch)
	router.HandleFunc("/tm-order/v1/customerapp/customers/verify", publicMiddleware.SetRouteChain(handler.Verify)).Methods(http.MethodGet)
	router.HandleFunc("/tm-order/v1/customerapp/customers/verify-change-email", publicMiddleware.SetRouteChain(handler.VerifyChangeEmail)).Methods(http.MethodGet)

	// SignUp(ctx context.Context, req SignUpRequest) (SignUpResponse, error)
	// SignIn(ctx context.Context, req SignInRequest) (SignInResponse, error)
	// SignOut(ctx context.Context) error
	// GetProfile(ctx context.Context) (GetProfileResponse, error)
	// UpdateProfile(ctx context.Context, req UpdateProfileRequest) error
	// ChangeEmail(ctx context.Context, req ChangeEmailRequest) (ChangeEmailResponse, error)
	// ChangePassword(ctx context.Context, req ChangePasswordRequest) error
	// Verify(ctx context.Context, req VerifyRequest) error
	// VerifyChangeEmail(ctx context.Context, req ChangeEmailVerificationRequest) error
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

func (handler HTTPHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req := SignUpRequest{}
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

	resp, err := handler.CustomerUseCase.SignUp(ctx, req)
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
		Message: "customer has been successfully signed up",
		Data:    resp,
	})
}

func (handler HTTPHandler) SignIn(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req := SignInRequest{}
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

	resp, err := handler.CustomerUseCase.SignIn(ctx, req)
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
		Message: "customer has been successfully signed in",
		Data:    resp,
	})
}

func (handler HTTPHandler) SignOut(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	err := handler.CustomerUseCase.SignOut(ctx)
	if err != nil {
		ae := errors.Destruct(err)
		response.JSON(w, ae.HTTPStatusCode, response.RESTEnvelope{
			Status:  ae.Status,
			Message: ae.Message,
		})

		return
	}

	response.JSON(w, http.StatusCreated, response.RESTEnvelope{
		Status:  status.OK,
		Message: "customer has been successfully signed out",
	})
}

func (handler HTTPHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	resp, err := handler.CustomerUseCase.GetProfile(ctx)
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
		Message: "customer profile",
		Data:    resp,
	})
}

func (handler HTTPHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req := UpdateProfileRequest{}
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

	err := handler.CustomerUseCase.UpdateProfile(ctx, req)
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
		Message: "customer has been successfully updated profile",
	})
}

func (handler HTTPHandler) ChangeEmail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req := ChangeEmailRequest{}
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

	resp, err := handler.CustomerUseCase.ChangeEmail(ctx, req)
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
		Message: "customer has been successfully requested an email change",
		Data:    resp,
	})
}

func (handler HTTPHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req := ChangePasswordRequest{}
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

	err := handler.CustomerUseCase.ChangePassword(ctx, req)
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
		Message: "customer has been successfully changed password",
	})
}

func (handler HTTPHandler) Verify(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req := VerifyRequest{}

	values := r.URL.Query()
	token := values.Get("token")

	req.Token = token

	err := handler.CustomerUseCase.Verify(ctx, req)
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
		Message: "customer verification succeded",
	})
}

func (handler HTTPHandler) VerifyChangeEmail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req := ChangeEmailVerificationRequest{}

	values := r.URL.Query()
	token := values.Get("token")

	req.Token = token

	err := handler.CustomerUseCase.VerifyChangeEmail(ctx, req)
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
		Message: "customer change email verification succeded",
	})
}
