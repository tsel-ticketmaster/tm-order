package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type HTTPRequestLogger interface {
	Middleware(http.Handler) http.Handler
}

type unimplementHTTPRequestLogger struct{}

func (m *unimplementHTTPRequestLogger) Middleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r)
	})
}

type httpRequestLoggerMiddleware struct {
	logger        *logrus.Logger
	httpStatusMap map[int]bool
}

func NewHTTPRequestLogger(logger *logrus.Logger, debug bool, statusCode ...int) HTTPRequestLogger {
	if debug {
		httpStatusMap := make(map[int]bool)
		for s := range statusCode {
			httpStatusMap[s] = true
		}
		return &httpRequestLoggerMiddleware{
			logger:        logger,
			httpStatusMap: httpStatusMap,
		}
	}

	return &unimplementHTTPRequestLogger{}
}

type wrappedResponseWriter struct {
	http.ResponseWriter
	recorder http.ResponseWriter
}

func (wrw wrappedResponseWriter) WriteHeader(statusCode int) {
	wrw.recorder.WriteHeader(statusCode)
	wrw.ResponseWriter.WriteHeader(statusCode)
}

func (wrw wrappedResponseWriter) Write(b []byte) (n int, err error) {
	wrw.recorder.Write(b)
	return wrw.ResponseWriter.Write(b)
}

func (m *httpRequestLoggerMiddleware) Middleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recoreder := httptest.NewRecorder()

		wrappedResponseWriter := wrappedResponseWriter{w, recoreder}

		requestHeader := r.Header

		buf, _ := io.ReadAll(r.Body)
		rcCopied1 := io.NopCloser(bytes.NewBuffer(buf))
		rcCopied2 := io.NopCloser(bytes.NewBuffer(buf))

		r.Body = rcCopied1

		now := time.Now()
		handler.ServeHTTP(wrappedResponseWriter, r)
		elapsed := time.Since(now)

		// requestBodyData := make(map[string]interface{})
		// json.NewDecoder(rcCopied2).Decode(&requestBodyData)
		var requestBodyData interface{}
		bindData(rcCopied2, &requestBodyData)

		result := recoreder.Result()

		defer result.Body.Close()

		var responseBodyData interface{}
		// json.NewDecoder(result.Body).Decode(&responseBodyData)
		bindData(result.Body, &responseBodyData)

		responseHeader := w.Header().Clone()

		captured := logrus.Fields{}
		captured["http.method"] = r.Method
		captured["http.url"] = r.RequestURI
		captured["http.request.body"] = requestBodyData
		captured["http.status_code"] = result.StatusCode
		for reqHeaderKey, reqHeaderCol := range requestHeader {
			captured[fmt.Sprintf("http.request.header.%s", strings.ReplaceAll(strings.ToLower(reqHeaderKey), " ", "_"))] = strings.Join(reqHeaderCol, ",")
		}
		for resHeaderKey, resHeaderCol := range responseHeader {
			captured[fmt.Sprintf("http.response.header.%s", strings.ReplaceAll(strings.ToLower(resHeaderKey), " ", "_"))] = strings.Join(resHeaderCol, ",")
		}
		captured["http.response.body"] = responseBodyData
		captured["time_consumption"] = elapsed.String()

		if ok := m.httpStatusMap[result.StatusCode]; ok {
			entry := m.logger.WithContext(r.Context()).WithFields(captured)
			entry.Info()
		}

	})
}

func bindData(r io.Reader, to any) {
	if to == nil {
		return
	}

	json.NewDecoder(r).Decode(to)
}
