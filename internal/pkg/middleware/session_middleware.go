package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/tsel-ticketmaster/tm-order/internal/pkg/jwt"
	"github.com/tsel-ticketmaster/tm-order/internal/pkg/session"
	"github.com/tsel-ticketmaster/tm-order/pkg/response"
	"github.com/tsel-ticketmaster/tm-order/pkg/status"
)

type AdminSession struct {
	jsonWebToken *jwt.JSONWebToken
	sess         session.Session
}

func NewAdminSessionMiddleware(jsonWebToken *jwt.JSONWebToken, sess session.Session) *AdminSession {
	return &AdminSession{
		jsonWebToken: jsonWebToken,
		sess:         sess,
	}
}

// Verify will verify the incomming request by checking authorization header.
func (s *AdminSession) Verify(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondUnauthorized(w, "invalid token")
			return
		}

		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 {
			respondUnauthorized(w, "invalid token")
			return
		}

		token := bearerToken[1]

		var claim jwt.Claim

		if err := s.jsonWebToken.Parse(ctx, token, &claim); err != nil {
			respondUnauthorized(w, err.Error())
			return
		}

		acc, err := s.sess.Get(ctx, claim.Subject)
		if err != nil {
			respondUnauthorized(w, err.Error())
			return
		}

		if acc.Type != "ADMIN" {
			respondUnauthorized(w, "invalid type of user")
			return
		}

		ctx = context.WithValue(ctx, session.AccountContextKey{}, acc)
		r = r.WithContext(ctx)

		next(w, r)
	}
}

func respondUnauthorized(w http.ResponseWriter, message string) {
	response.JSON(w, http.StatusUnauthorized, response.RESTEnvelope{
		Status:  status.UNAUTHORIZED,
		Message: message,
	})
}

type CustomerSession struct {
	jsonWebToken *jwt.JSONWebToken
	sess         session.Session
}

func NewCustomerSessionMiddleware(jsonWebToken *jwt.JSONWebToken, sess session.Session) *CustomerSession {
	return &CustomerSession{
		jsonWebToken: jsonWebToken,
		sess:         sess,
	}
}

// Verify will verify the incomming request by checking authorization header.
func (s *CustomerSession) Verify(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondUnauthorized(w, "invalid token")
			return
		}

		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 {
			respondUnauthorized(w, "invalid token")
			return
		}

		token := bearerToken[1]

		var claim jwt.Claim

		if err := s.jsonWebToken.Parse(ctx, token, &claim); err != nil {
			respondUnauthorized(w, err.Error())
			return
		}

		acc, err := s.sess.Get(ctx, claim.Subject)
		if err != nil {
			respondUnauthorized(w, err.Error())
			return
		}

		if acc.Type != "CUSTOMER" {
			respondUnauthorized(w, "invalid type of user")
			return
		}

		ctx = context.WithValue(ctx, session.AccountContextKey{}, acc)
		r = r.WithContext(ctx)

		next(w, r)
	}
}
