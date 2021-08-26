package web

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
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
	if err := r.ParseForm(); err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.FormValue("username") != "admin" || r.FormValue("password") != "admin" {
		http.Error(w, "unauthorised", http.StatusUnauthorized)
	}

	http.Redirect(w, r, "/main.html", http.StatusTemporaryRedirect)
}
