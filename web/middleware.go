package web

import (
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

type MiddlewareFunc func(h http.HandlerFunc) http.HandlerFunc

func Decorate(h http.HandlerFunc, decorators ...MiddlewareFunc) http.HandlerFunc {
	var rootHandler = CORS(h)
	for _, decorator := range decorators {
		rootHandler = decorator(rootHandler)
	}
	return rootHandler
}

func LoggingMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.WithFields(log.Fields{"endpoint": strings.Replace(r.URL.Path, "/api/", "", -1), "method": r.Method, "headers": sanitiseHeaders(r.Header)}).Info("api request:")
		h(w, r)
	}
}

func CORS(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.Header().Add("Access-Control-Allow-Origin", "*")
			w.WriteHeader(http.StatusOK)
			return
		}
		h(w, r)
	}
}

func sanitiseHeaders(headers http.Header) http.Header {
	strippedHeaders := make(http.Header)
	for key, val := range headers {
		if !strings.EqualFold(key, "authorization") {
			strippedHeaders[key] = val
		}
	}

	return strippedHeaders
}
