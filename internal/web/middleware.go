package web

import (
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/web/security"
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

func Secure(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claim := r.Header.Get("authorization")
		if len(claim) == 0 {
			log.Debug("rejecting empty claim")
			http.Error(w, "unauthorised", http.StatusUnauthorized)
			return
		}

		token, err := security.ParseClaim(claim)
		if err != nil {
			log.Error("token parse claim error: ", err.Error())
			http.Error(w, "unauthorised", http.StatusUnauthorized)
			return
		}

		if time.Now().After(token.Expires) {
			log.Debug("token expired")
			http.Error(w, "unauthorised", http.StatusUnauthorized)
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
