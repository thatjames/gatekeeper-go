package web

import (
	"embed"
	"encoding/json"
	"io/fs"
	"net/http"
	"strings"

	"gitlab.com/thatjames-go/gatekeeper-go/web/domain"
)

//go:embed ui
var efs embed.FS

func Init() error {
	fsys, err := fs.Sub(efs, "ui")
	if err != nil {
		return err
	}

	fs := http.FileServer(http.FS(fsys))
	http.HandleFunc("/api/login", makeEndpoint(http.MethodPost, login, LoggingMiddleware))
	http.Handle("/", fs)
	if err := http.ListenAndServe(":5000", nil); err != nil {
		return err
	}
	return nil
}

func makeEndpoint(method string, handler http.HandlerFunc, decorators ...MiddlewareFunc) http.HandlerFunc {
	h := func(w http.ResponseWriter, r *http.Request) {
		if !strings.EqualFold(method, r.Method) {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		handler(w, r)
	}

	return Decorate(h, decorators...)
}

//start of handlers

func login(w http.ResponseWriter, r *http.Request) {
	var req domain.UserLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "malformed request", http.StatusBadRequest)
		return
	}

	if req.Username != "admin" || req.Password != "admin" {
		http.Error(w, "forbidden", http.StatusUnauthorized)
		return
	}
}
