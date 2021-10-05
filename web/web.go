package web

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"path"
	"strings"
	"text/template"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/tg123/go-htpasswd"
	"gitlab.com/thatjames-go/gatekeeper-go/config"
	"gitlab.com/thatjames-go/gatekeeper-go/dhcp"
	"gitlab.com/thatjames-go/gatekeeper-go/system"
	"gitlab.com/thatjames-go/gatekeeper-go/web/domain"
	"gitlab.com/thatjames-go/gatekeeper-go/web/security"
)

//go:embed ui
var efs embed.FS

//go:embed ui/pages/dhcp.html
var dhcpTempl string

//go:embed ui/pages/home.html
var homeTempl string

var (
	leaseDB *dhcp.LeaseDB
	version string
)

type PageFs struct {
	fsys fs.FS
}

func (p PageFs) Open(name string) (fs.File, error) {
	if name == "main" {
		name += ".html"
	}
	return p.fsys.Open(name)
}

func Init(ver string, config *config.Web, leases *dhcp.LeaseDB) error {
	version = ver
	var err error
	fsys, err := fs.Sub(efs, "ui")
	if err != nil {
		return err
	}

	leaseDB = leases
	fs := http.FileServer(http.FS(PageFs{fsys}))
	http.Handle("/", fs)
	http.HandleFunc("/page/", makeEndpoint(http.MethodGet, templateHandler, Secure))
	http.HandleFunc("/api/login", makeEndpoint(http.MethodPost, login, LoggingMiddleware))
	http.HandleFunc("/api/version", makeEndpoint(http.MethodGet, getVersion, Secure))
	if config.TLS != nil {
		log.Info("Start TLS listener on ", config.Address)
		if err := http.ListenAndServeTLS(config.Address, config.TLS.PublicKey, config.TLS.PrivateKey, nil); err != nil {
			return err
		}
	} else {
		log.Info("Start clear-text listener on ", config.Address)
		if err := http.ListenAndServe(config.Address, nil); err != nil {
			return err
		}
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

	if passwd, err := htpasswd.New(config.Config.Web.HTPasswdFile, htpasswd.DefaultSystems, nil); err == nil {
		if !passwd.Match(req.Username, req.Password) {
			http.Error(w, "forbidden", http.StatusUnauthorized)
			return
		}
	} else {
		log.Warn("unable to read htpasswd file: ", err.Error())
		log.Warn("defaulting to default username/password")
		if req.Username != "admin" || req.Password != "admin" {
			http.Error(w, "forbidden", http.StatusUnauthorized)
			return
		}
	}

	authToken, err := security.CreateAuthToken(req.Username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response := map[string]string{
		"token": authToken,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getVersion(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, version)
}

func templateHandler(w http.ResponseWriter, r *http.Request) {
	page := template.New("main").Funcs(template.FuncMap{"Format": format})
	var (
		pageData interface{}
		err      error
	)
	switch strings.ToLower(path.Base(r.URL.Path)) {
	case "dhcp":
		pageData = LeasePage{
			ReservedLeases: leaseDB.ReservedLeases(),
			ActiveLeases:   leaseDB.ActiveLeases(),
			Start:          config.Config.DHCP.StartAddr,
			End:            config.Config.DHCP.EndAddr,
			Nameservers:    config.Config.DHCP.NameServers,
			DomainName:     config.Config.DHCP.DomainName,
		}
		if page, err = page.Parse(dhcpTempl); err != nil {
			http.Error(w, "failed", http.StatusInternalServerError)
			fmt.Println(err.Error())
			return
		}

	case "home":
		if pageData, err = system.GetSystemInfo(); err != nil {
			fmt.Fprintln(w, "<div>failed: "+err.Error()+"</div>")
		}
		if page, err = page.Parse(homeTempl); err != nil {
			http.Error(w, "failed", http.StatusInternalServerError)
			fmt.Println(err.Error())
			return
		}

	}
	if err = page.Execute(w, pageData); err != nil {
		http.Error(w, "failed", http.StatusInternalServerError)
		fmt.Println(err.Error())
		return
	}
}

func format(t time.Time) string {
	if t.IsZero() {
		return " - "
	}
	return t.Format("2006-01-02 15:04:05")
}
