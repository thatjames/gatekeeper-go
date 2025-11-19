package dns

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/config"
	"gitlab.com/thatjames-go/gatekeeper-go/internal/util"
)

type DNSServerOpts struct {
	Options       *DNSServerOpts
	Interface     string
	Port          int
	Upstream      []string
	ResolverOpts  *ResolverOpts
	BlocklistUrls []string
}

var defaultDNSServerOpts = DNSServerOpts{
	Interface: "eth0",
	Port:      53,
}

type DNSServer struct {
	opts         *DNSServerOpts
	resolver     *DNSResolver
	packetConn   net.PacketConn
	receiverChan chan *dnsWorkItem
	responseChan chan *dnsWorkItem
	exitChan     chan struct{}
	lock         sync.Mutex
}

type dnsWorkItem struct {
	*DNSPacket
	err       error
	startTime time.Time
}

func NewDNSServer() *DNSServer {
	return NewDNSServerWithOpts(defaultDNSServerOpts)
}

func NewDNSServerWithOpts(opts DNSServerOpts) *DNSServer {
	return &DNSServer{
		resolver:     NewDNSResolverWithOpts(*opts.ResolverOpts),
		opts:         &opts,
		packetConn:   nil,
		receiverChan: make(chan *dnsWorkItem, 100),
		responseChan: make(chan *dnsWorkItem, 100),
		lock:         sync.Mutex{},
	}
}

func (d *DNSServer) Start() error {
	log.Info("starting DNS server")
	log.Tracef("starting DNS server on port %d", d.opts.Port)
	var err error
	if d.packetConn, err = net.ListenPacket("udp4", fmt.Sprintf(":%d", d.opts.Port)); err != nil {
		log.Error("unable to start DNS server: ", err.Error())
		return err
	}
	d.exitChan = make(chan struct{})
	go d.listen()
	go d.receiverWorker()
	go d.responseWorker()
	if len(d.opts.BlocklistUrls) > 0 {
		d.LoadBlocklistFromURLS(d.opts.BlocklistUrls)
	}
	log.Info("DNS server started")
	return nil
}

func (d *DNSServer) Stop() error {
	close(d.exitChan)
	log.Info("stopping DNS server")
	d.lock.Lock() //wait for the goroutines to finish
	defer d.lock.Unlock()
	return nil
}

func (d *DNSServer) Options() *DNSServerOpts {
	return d.opts
}

func (d *DNSServer) UpdateOptions(opts *DNSServerOpts) {
	d.opts = opts
}

func (d *DNSServer) AddLocalDomain(domain string, ip string) error {
	return d.resolver.AddLocalDomain(domain, net.ParseIP(ip).To4())
}

func (d *DNSServer) DeleteLocalDomain(domain string) {
	d.resolver.DeleteLocalDomain(domain)
}

func (d *DNSServer) FlushBlocklist() {
	d.resolver.FlushBlocklist()
}

func (d *DNSServer) AddBlocklistFromURL(url string) error {
	var dat []byte
	var err error
	if strings.HasPrefix(url, "http") {
		resp, err := http.DefaultClient.Get(url)
		if err != nil {
			log.Warnf("unable to fetch blocklist %s: %s", url, err.Error())
			return err
		}
		defer resp.Body.Close()
		dat, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Warnf("unable to read blocklist %s: %s", url, err.Error())
			return err
		}
	} else {
		log.Debug("loading blocklist from file: ", url)
		dat, err = ioutil.ReadFile(url)
		if err != nil {
			log.Warnf("unable to read blocklist %s: %s", url, err.Error())
			return err
		}
	}
	if hosts, ok := util.ValidateIsHostFileFormat(string(dat)); ok {
		d.resolver.AddBlocklistEntries(hosts)
	} else {
		return errors.New("invalid blocklist format")
	}
	return nil
}

func (d *DNSServer) LoadBlocklistFromURLS(urls []string) {
	log.Debugf("loading blocklist from URLs %v", urls)
	blockedDomains := make([]string, 0)
	http.DefaultClient.Timeout = time.Second * 15
	resultChan := make(chan []string, len(config.Config.DNS.BlockLists))
	signalChan := make(chan interface{}, len(config.Config.DNS.BlockLists))
	var workerCount = len(urls)
	for _, urlToLoad := range urls {
		go func() {
			defer func() {
				signalChan <- nil
			}()
			var dat []byte
			var err error
			if strings.HasPrefix(urlToLoad, "http") {
				resp, err := http.DefaultClient.Get(urlToLoad)
				if err != nil {
					log.Warnf("unable to fetch blocklist %s: %s", urlToLoad, err.Error())
					return
				}
				defer resp.Body.Close()
				dat, err = ioutil.ReadAll(resp.Body)
				if err != nil {
					log.Warnf("unable to read blocklist %s: %s", urlToLoad, err.Error())
					return
				}
			} else {
				log.Debug("loading blocklist from file: ", urlToLoad)
				dat, err = ioutil.ReadFile(urlToLoad)
				if err != nil {
					log.Warnf("unable to read blocklist %s: %s", urlToLoad, err.Error())
					return
				}
			}
			if hosts, ok := util.ValidateIsHostFileFormat(string(dat)); ok {
				resultChan <- hosts
			}
		}()
	}
	for workerCount > 0 {
		select {
		case <-signalChan:
			workerCount--

		case results := <-resultChan:
			blockedDomains = append(blockedDomains, results...)
		}
	}
	close(signalChan)
	close(resultChan)
	log.Infof("loaded %d blocked domains", len(blockedDomains))
	d.resolver.AddBlocklistEntries(blockedDomains)
}

func (d *DNSServer) listen() {
	d.lock.Lock()
	defer func() {
		log.Tracef("closing DNS server")
		d.packetConn.Close()
		d.lock.Unlock()
	}()
	for {
		select {
		case <-d.exitChan:
			return
		default:
			buff := make([]byte, 1500)
			d.packetConn.SetReadDeadline(time.Now().Add(time.Second * 2))
			n, addr, err := d.packetConn.ReadFrom(buff)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					// This is a timeout error, don't log it
					continue
				}
				log.Error("unable to read datastream: ", err.Error())
				continue
			}
			workItem := &dnsWorkItem{
				startTime: time.Now(),
			}
			buff = buff[:n]
			log.Tracef("received %d bytes from %s", n, addr.String())
			msg, err := ParseDNSMessage(buff)
			if err != nil {
				log.Error("unable to parse DNS message: ", err.Error())
				workItem.err = err
			} else if msg.Header == nil {
				log.Error("unable to parse DNS message: no header")
			} else {
				workItem.DNSPacket = &DNSPacket{
					DNSMessage:   msg,
					ResponseAddr: addr,
				}
			}
			d.receiverChan <- workItem
		}
	}
}

func (d *DNSServer) receiverWorker() {
	for packet := range d.receiverChan {
		if packet.err != nil {
			log.Errorf("skipping malformed packet: %v", packet.err)
			continue
		}

		if packet.DNSMessage == nil || packet.DNSMessage.Header == nil {
			log.Error("received packet with nil DNSMessage or Header")
			continue
		}

		log.Tracef("received DNS packet from %s", packet.ResponseAddr.String())
		responses, authorities, err := d.resolver.Resolve(packet.DNSMessage.Questions[0].ParsedName, packet.DNSMessage.Questions[0].Type)

		if err != nil {
			if err == ErrNxDomain {
				packet.DNSMessage.Header.SetRCODE(RCODENameFailure)
			} else {
				packet.DNSMessage.Header.SetRCODE(RCODEServerFailure)
			}
		} else {
			packet.DNSMessage.Header.SetRCODE(RCODESuccess)
			if responses != nil {
				packet.DNSMessage.Answers = responses
				log.Tracef("adding answer %v", responses)
			}
			if authorities != nil {
				packet.DNSMessage.Authorities = authorities
				log.Tracef("adding authority %s", authorities)
			}
		}

		d.responseChan <- packet

		// we do this afterwards to not interfere with the response timing
		if packet.err != nil {
			queryByIPCounter.With(prometheus.Labels{"ip": strings.Split(packet.ResponseAddr.String(), ":")[0], "result": "failed"}).Inc()
		} else {
			queryByIPCounter.With(prometheus.Labels{"ip": strings.Split(packet.ResponseAddr.String(), ":")[0], "result": "success"}).Inc()
		}
	}
}

func (d *DNSServer) responseWorker() {
	for packet := range d.responseChan {
		log.Tracef("sending DNS response packet to %s", packet.ResponseAddr.String())
		packet.Header.SetQR(true)
		packet.Header.SetRA(true)
		packet.DNSMessage.Additionals = nil
		data, err := MarshalDNSMessage(packet.DNSMessage)
		if err != nil {
			log.Error("unable to marshal DNS packet: ", err.Error())
			continue
		}
		n, err := d.packetConn.WriteTo(data, packet.ResponseAddr)
		log.Tracef("sent %d bytes to %s", n, packet.ResponseAddr.String())
		timeElapsed := time.Since(packet.startTime).Round(time.Millisecond)
		reqDuration.Observe(float64(timeElapsed.Milliseconds()))
		if err != nil {
			log.Error("unable to send DNS packet: ", err.Error())
		}
	}
}
