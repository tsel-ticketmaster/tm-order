package middleware

import "net/http"

type InternalService interface {
	Verify(http.HandlerFunc) http.HandlerFunc
}
