package web

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path"
	"strings"
	"text/template"
	"time"

	"gitlab.com/thatjames-go/gatekeeper-go/config"
	"gitlab.com/thatjames-go/gatekeeper-go/dhcp"
	"gitlab.com/thatjames-go/gatekeeper-go/util"
	"gitlab.com/thatjames-go/gatekeeper-go/web/domain"
	"golang.org/x/sys/unix"
)

//go:embed ui
var efs embed.FS

//go:embed ui/pages/dhcp.html
var dhcpTempl string

//go:embed ui/pages/home.html
var homeTempl string

var (
	leaseDB *dhcp.LeaseDB
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

func Init(config *config.Web, leases *dhcp.LeaseDB) error {
	var err error
	fsys, err := fs.Sub(efs, "ui")
	if err != nil {
		return err
	}

	leaseDB = leases
	fs := http.FileServer(http.FS(PageFs{fsys}))
	http.Handle("/", fs)
	http.HandleFunc("/page/", templateHandler)
	http.HandleFunc("/api/login", makeEndpoint(http.MethodPost, login, LoggingMiddleware))
	if err := http.ListenAndServe(config.Address, nil); err != nil {
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
		var t unix.Sysinfo_t
		if err := unix.Sysinfo(&t); err != nil {
			http.Error(w, "failed", http.StatusInternalServerError)
			fmt.Println(err.Error())
			return
		}
		hostname, _ := os.Hostname()
		pageData = HomePage{
			Hostname: hostname,
			Uptime:   (time.Second * time.Duration(t.Uptime)).Round(time.Second).String(),
			Freeram:  util.ByteSize(int(t.Freeram)),
			Totalram: util.ByteSize(int(t.Totalram)),
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
