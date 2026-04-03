package dns

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/util"
)

type BlocklistFetcher interface {
	Fetch(url string) ([]string, error)
}

type HTTPBlocklistFetcher struct {
	client *http.Client
}

func NewHTTPBlocklistFetcher() *HTTPBlocklistFetcher {
	return &HTTPBlocklistFetcher{
		client: &http.Client{
			Timeout: time.Second * 15,
		},
	}
}

func (f *HTTPBlocklistFetcher) Fetch(url string) ([]string, error) {
	var dat []byte
	var err error

	if strings.HasPrefix(url, "http") {
		resp, err := f.client.Get(url)
		if err != nil {
			log.Warnf("unable to fetch blocklist %s: %s", url, err.Error())
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			log.Warnf("unable to fetch blocklist %s: HTTP %d", url, resp.StatusCode)
			return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
		}
		dat, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Warnf("unable to read blocklist %s: %s", url, err.Error())
			return nil, err
		}
	} else {
		log.Debug("loading blocklist from file: ", url)
		dat, err = ioutil.ReadFile(url)
		if err != nil {
			log.Warnf("unable to read blocklist %s: %s", url, err.Error())
			return nil, err
		}
	}

	hosts, ok := util.ValidateIsHostFileFormat(string(dat))
	if !ok {
		return nil, ErrInvalidBlocklistFormat
	}

	return hosts, nil
}

type FileBlocklistFetcher struct{}

func (f *FileBlocklistFetcher) Fetch(url string) ([]string, error) {
	dat, err := ioutil.ReadFile(url)
	if err != nil {
		log.Warnf("unable to read blocklist %s: %s", url, err.Error())
		return nil, err
	}

	hosts, ok := util.ValidateIsHostFileFormat(string(dat))
	if !ok {
		return nil, ErrInvalidBlocklistFormat
	}

	return hosts, nil
}
