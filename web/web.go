package web

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"strings"
	"time"

	"gitlab.com/thatjames-go/gatekeeper-go/config"
	"gitlab.com/thatjames-go/gatekeeper-go/dhcp"
	"gitlab.com/thatjames-go/gatekeeper-go/web/domain"
)

//go:embed ui
var efs embed.FS

//go:embed templates/main.tmpl
var mainTempl string

var (
	mainTemplate = template.Must(template.New("main").Funcs(template.FuncMap{"Format": format}).Parse(mainTempl))
	leaseDB      *dhcp.LeaseDB
)

func Init(db *dhcp.LeaseDB) error {
	var err error
	leaseDB = db
	fsys, err := fs.Sub(efs, "ui")
	if err != nil {
		return err
	}

	fs := http.FileServer(http.FS(fsys))
	http.Handle("/", fs)
	http.HandleFunc("/main", templateHandler)
	http.HandleFunc("/api/login", makeEndpoint(http.MethodPost, login, LoggingMiddleware))
	if err := http.ListenAndServe(config.Config.Web.Address, nil); err != nil {
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

func templateHandler(w http.ResponseWriter, r *http.Request) {
	if err := mainTemplate.Execute(w, leaseDB.ActiveLeases()); err != nil {
		http.Error(w, "failed", http.StatusInternalServerError)
		fmt.Println(err.Error())
		return
	}
}

func format(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}
