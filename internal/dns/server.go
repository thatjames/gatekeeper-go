package dns

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

type DNSServerOpts struct {
	Options        *DNSServerOpts
	Interface      string
	Port           int
	Upstream       []string
	ResolverOpts   *ResolverOpts
	BlocklistUrls  []string
	BlockedDomains []string
	ReadDeadline   time.Duration
}

// tcpIdleTimeout closes a TCP connection that sends no query for this long.
const tcpIdleTimeout = time.Second * 10

var defaultDNSServerOpts = DNSServerOpts{
	Interface:    "eth0",
	Port:         53,
	Upstream:     []string{"8.8.8.8", "1.1.1.1"},
	ReadDeadline: time.Second * 2,
}

type DNSServer struct {
	opts             *DNSServerOpts
	resolver         Resolver
	blocklistFetcher BlocklistFetcher
	packetConn       net.PacketConn
	tcpListener      net.Listener
	receiverChan     chan *dnsWorkItem
	responseChan     chan *dnsWorkItem
	exitChan         chan struct{}
	lock             sync.Mutex
	readDeadline     time.Duration
}

type dnsWorkItem struct {
	*DNSPacket
	err       error
	startTime time.Time
}

func NewDNSServer() *DNSServer {
	return NewDNSServerWithOpts(defaultDNSServerOpts, nil, nil)
}

// resolverOptsFrom derives the resolver options from the server options,
// falling back to the server's upstreams when the caller didn't supply any.
func resolverOptsFrom(opts *DNSServerOpts) ResolverOpts {
	resolverOpts := ResolverOpts{}
	if opts.ResolverOpts != nil {
		resolverOpts = *opts.ResolverOpts
	}
	if len(resolverOpts.Upstreams) == 0 {
		resolverOpts.Upstreams = opts.Upstream
	}
	if resolverOpts.LocalDomains == nil {
		resolverOpts.LocalDomains = make(map[string]net.IP)
	}
	return resolverOpts
}

func NewDNSServerWithOpts(opts DNSServerOpts, resolver Resolver, fetcher BlocklistFetcher) *DNSServer {
	if resolver == nil {
		resolver = NewDNSResolverWithOpts(resolverOptsFrom(&opts))
	}
	if fetcher == nil {
		fetcher = NewHTTPBlocklistFetcher()
	}

	readDeadline := opts.ReadDeadline
	if readDeadline == 0 {
		readDeadline = time.Second * 2
	}

	return &DNSServer{
		opts:             &opts,
		resolver:         resolver,
		blocklistFetcher: fetcher,
		packetConn:       nil,
		receiverChan:     make(chan *dnsWorkItem, 100),
		responseChan:     make(chan *dnsWorkItem, 100),
		lock:             sync.Mutex{},
		readDeadline:     readDeadline,
	}
}

func (d *DNSServer) Start() error {
	if d.resolver == nil {
		log.Warn("resolver was nil, creating new one with upstreams: ", d.opts.Upstream)
		d.resolver = NewDNSResolverWithOpts(resolverOptsFrom(d.opts))
	}
	if len(d.opts.BlockedDomains) > 0 {
		log.Debugf("adding %d blocked domains", len(d.opts.BlockedDomains))
		d.resolver.AddBlocklistEntries(d.opts.BlockedDomains)
	}
	log.Info("starting DNS server")
	log.Tracef("starting DNS server on port %d", d.opts.Port)
	var err error
	listenAddr := fmt.Sprintf(":%d", d.opts.Port)
	if d.packetConn, err = net.ListenPacket("udp4", listenAddr); err != nil {
		log.Error("unable to start DNS server: ", err.Error())
		return err
	}
	if d.tcpListener, err = net.Listen("tcp4", listenAddr); err != nil {
		log.Error("unable to start DNS server over TCP: ", err.Error())
		d.packetConn.Close()
		return err
	}
	d.exitChan = make(chan struct{})
	go d.listen()
	go d.listenTCP()
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
	if d.tcpListener != nil {
		d.tcpListener.Close()
	}
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
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return fmt.Errorf("invalid IP address: %s", ip)
	}
	return d.resolver.AddLocalDomain(domain, parsedIP.To4())
}

func (d *DNSServer) DeleteLocalDomain(domain string) {
	d.resolver.DeleteLocalDomain(domain)
}

func (d *DNSServer) FlushBlocklist() {
	d.resolver.FlushBlocklist()
}

func (d *DNSServer) AddBlocklistFromURL(url string) error {
	hosts, err := d.blocklistFetcher.Fetch(url)
	if err != nil {
		return err
	}
	d.resolver.AddBlocklistEntries(hosts)
	return nil
}

func (d *DNSServer) LoadBlocklistFromURLS(urls []string) {
	log.Debugf("loading blocklist from URLs %v", urls)
	blockedDomains := make([]string, 0)
	resultChan := make(chan []string, len(urls))
	signalChan := make(chan struct{}, len(urls))
	var workerCount = len(urls)
	for _, urlToLoad := range urls {
		go func() {
			defer func() {
				signalChan <- struct{}{}
			}()
			hosts, err := d.blocklistFetcher.Fetch(urlToLoad)
			if err != nil {
				log.Warnf("unable to fetch blocklist %s: %s", urlToLoad, err.Error())
				return
			}
			resultChan <- hosts
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

func (d *DNSServer) DeleteBlockedDomain(domain string) {
	d.resolver.DeleteBlocklistEntry(domain)
}

func (d *DNSServer) AddBlockedDomain(domain string) {
	d.resolver.AddBlocklistEntries([]string{domain})
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
			d.packetConn.SetReadDeadline(time.Now().Add(d.readDeadline))
			n, addr, err := d.packetConn.ReadFrom(buff)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					log.Tracef("timeout error, no packet")
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
			log.Tracef("push packet to worker")
			d.receiverChan <- workItem
			log.Tracef("packet pushed to worker")
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
		d.answer(packet.DNSMessage)

		log.Tracef("push packet to response worker")
		d.responseChan <- packet
		log.Tracef("packet pushed to response worker")

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
		data, err := prepareReply(packet.DNSMessage)
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

// answer resolves the first question of msg and fills msg in as the reply:
// RCODE, answers and authorities.
func (d *DNSServer) answer(msg *DNSMessage) {
	responses, authorities, err := d.resolver.Resolve(msg.Questions[0].ParsedName, msg.Questions[0].Type)
	if err != nil {
		if err == ErrNxDomain {
			msg.Header.SetRCODE(RCODENameFailure)
		} else {
			msg.Header.SetRCODE(RCODEServerFailure)
		}
		return
	}
	msg.Header.SetRCODE(RCODESuccess)
	if responses != nil {
		msg.Answers = responses
		log.Tracef("adding answer %v", responses)
	}
	if authorities != nil {
		msg.Authorities = authorities
		log.Tracef("adding authority %s", authorities)
	}
}

// prepareReply sets the reply flags on an answered query and marshals it.
func prepareReply(msg *DNSMessage) ([]byte, error) {
	msg.Header.SetQR(true)
	msg.Header.SetRA(true)
	msg.Additionals = nil
	return MarshalDNSMessage(msg)
}

func (d *DNSServer) listenTCP() {
	for {
		conn, err := d.tcpListener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				log.Tracef("closing DNS TCP listener")
				return
			}
			log.Error("unable to accept TCP connection: ", err.Error())
			continue
		}
		go d.serveTCP(conn)
	}
}

// serveTCP answers length-prefixed queries on one connection, in order,
// until the client goes quiet, sends something unreadable or the server stops.
func (d *DNSServer) serveTCP(conn net.Conn) {
	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-d.exitChan:
		case <-done:
		}
		conn.Close()
	}()

	clientIP := strings.Split(conn.RemoteAddr().String(), ":")[0]
	lengthPrefix := make([]byte, 2)
	for {
		conn.SetReadDeadline(time.Now().Add(tcpIdleTimeout))
		if _, err := io.ReadFull(conn, lengthPrefix); err != nil {
			return
		}
		buff := make([]byte, binary.BigEndian.Uint16(lengthPrefix))
		if _, err := io.ReadFull(conn, buff); err != nil {
			return
		}
		startTime := time.Now()
		msg, err := ParseDNSMessage(buff)
		if err != nil {
			log.Error("unable to parse DNS message over TCP: ", err.Error())
			return
		}
		if len(msg.Questions) == 0 {
			log.Error("DNS message over TCP has no question")
			return
		}

		log.Tracef("received DNS packet over TCP from %s", clientIP)
		d.answer(msg)
		data, err := prepareReply(msg)
		if err != nil {
			log.Error("unable to marshal DNS packet: ", err.Error())
			return
		}
		reply := make([]byte, 2, 2+len(data))
		binary.BigEndian.PutUint16(reply, uint16(len(data)))
		reply = append(reply, data...)
		conn.SetWriteDeadline(time.Now().Add(tcpIdleTimeout))
		_, err = conn.Write(reply)
		reqDuration.Observe(float64(time.Since(startTime).Round(time.Millisecond).Milliseconds()))
		queryByIPCounter.With(prometheus.Labels{"ip": clientIP, "result": "success"}).Inc()
		if err != nil {
			log.Error("unable to send DNS packet over TCP: ", err.Error())
			return
		}
	}
}
