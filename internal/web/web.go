package web

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"text/template"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/tg123/go-htpasswd"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/config"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/dhcp"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/system"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/web/domain"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/web/security"
)

//go:embed ui
var efs embed.FS

var (
	leaseDB *dhcp.LeaseDB
	version string
)

func Init(ver string, config *config.Web, leases *dhcp.LeaseDB) error {
	version = ver
	leaseDB = leases
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", makeEndpoint(http.MethodGet, pageRender))
	http.HandleFunc("/api/verify", makeEndpoint(http.MethodGet, verify, Secure))
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

func verify(w http.ResponseWriter, r *http.Request) {
	//secured endpoint, middleware decides what happens here
}

func pageRender(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Path
	switch {
	case name == "/", name == "", name == ".":
		name = "index.html"
	case !strings.HasPrefix(name, "/assets"):
		name = fmt.Sprintf("page/%s", path.Base(name))
	}
	name = path.Join("ui", name)
	switch {
	case strings.HasPrefix(name, "ui/assets"), name == "ui/index.html":
		if err := loadStaticFile(w, name); err != nil {
			if os.IsNotExist(err) {
				http.Error(w, "not found", http.StatusNotFound)
			} else {
				http.Error(w, "general failure", http.StatusInternalServerError)
			}
		}

	default:
		w.Header().Add("Content-Type", "text/html")
		if err := renderPage(w, name); err != nil {
			http.Error(w, "render failed: "+err.Error(), http.StatusInternalServerError)
		}
	}
}

func loadStaticFile(w http.ResponseWriter, name string) error {
	f, err := efs.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()
	var mimeType string
	extension := strings.Split(path.Base(name), ".")[1]
	switch extension {
	case "js":
		mimeType = "text/javascript"
	case "ico":
		mimeType = "image/x-icon"
	default:
		mimeType = "text/" + extension
	}
	w.Header().Add("Content-Type", mimeType)
	io.Copy(w, f)
	return nil
}

func renderPage(w io.Writer, name string) error {
	rawDat, err := efs.ReadFile(fmt.Sprintf("%s.html", name))
	if err != nil {
		return err
	}

	page, err := template.New(path.Base(name)).Funcs(template.FuncMap{"Format": format}).Parse(string(rawDat))
	if err != nil {
		return err
	}

	var pageData interface{}

	switch strings.ToLower(path.Base(name)) {
	case "dhcp":
		pageData = LeasePage{
			ReservedLeases: leaseDB.ReservedLeases(),
			ActiveLeases:   leaseDB.ActiveLeases(),
			Start:          config.Config.DHCP.StartAddr,
			End:            config.Config.DHCP.EndAddr,
			Nameservers:    config.Config.DHCP.NameServers,
			DomainName:     config.Config.DHCP.DomainName,
		}

	case "home":
		if pageData, err = system.GetSystemInfo(); err != nil {
			fmt.Fprintln(w, "<div>failed: "+err.Error()+"</div>")
		}
	}

	return page.Execute(w, pageData)
}

func format(t time.Time) string {
	if t.IsZero() {
		return " - "
	}
	return t.Format("2006-01-02 15:04:05")
}
