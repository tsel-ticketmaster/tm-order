package middleware

import "net/http"

func SetChain(router http.Handler, middleware ...func(http.Handler) http.Handler) http.Handler {
	var handler http.Handler = router

	if len(middleware) < 1 {
		return handler
	}

	for i := len(middleware) - 1; i >= 0; i-- {
		handler = middleware[i](handler)
	}

	return handler
}

func SetRouteChain(handler http.HandlerFunc, next ...func(http.HandlerFunc) http.HandlerFunc) http.HandlerFunc {
	if len(next) < 1 {
		return handler
	}

	for i := len(next) - 1; i >= 0; i-- {
		handler = next[i](handler)
	}

	return handler
}
