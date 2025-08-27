package dhcp

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type LeaseState int

func (ls LeaseState) String() string {
	switch ls {
	case LeaseAvailable:
		return "Available"
	case LeaseOffered:
		return "Offered"
	case LeaseReserved:
		return "Reserved"
	case LeaseActive:
		return "Active"
	default:
		return "unknown"
	}
}

const (
	LeaseAvailable LeaseState = iota
	LeaseOffered
	LeaseReserved
	LeaseActive
)

var ()

type Lease struct {
	ClientId string
	Hostname string
	IP       net.IP
	Expiry   time.Time
	State    LeaseState
}

func (l *Lease) String() string {
	return fmt.Sprintf("%s: %s - %s: %s expiring at %s", l.Hostname, l.IP.String(), l.ClientId, l.State, l.Expiry.Format("15:04:05"))
}

func (l *Lease) Clear() {
	l.ClientId = ""
	l.Hostname = ""
	l.Expiry = time.Time{}
	l.State = LeaseAvailable
}

type LeaseDB struct {
	start             net.IP
	end               net.IP
	leases            []*Lease
	lock              *sync.Mutex
	reservedAddresses map[string]*Lease
}

func NewLeaseDB(startAddr, endAddr net.IP) *LeaseDB {
	leaseRange := int(binary.BigEndian.Uint32(endAddr) - binary.BigEndian.Uint32(startAddr))
	start := binary.BigEndian.Uint32(startAddr)
	leases := make([]*Lease, leaseRange)
	for i := range leases {
		leases[i] = &Lease{
			IP: make(net.IP, 4),
		}
		binary.BigEndian.PutUint32(leases[i].IP, start+uint32(i))
	}
	return &LeaseDB{
		start:             startAddr,
		end:               endAddr,
		lock:              new(sync.Mutex),
		leases:            leases,
		reservedAddresses: make(map[string]*Lease),
	}
}

func (l *LeaseDB) GetLease(clientId string) *Lease {
	l.lock.Lock()
	defer l.lock.Unlock()
	if reservedLease, ok := l.reservedAddresses[clientId]; ok {
		return reservedLease
	}
	for _, lease := range l.leases {
		if strings.EqualFold(clientId, lease.ClientId) {
			if time.Now().After(lease.Expiry) {
				lease.Clear()
				return nil
			}
			return lease
		}
	}
	return nil
}

func (l *LeaseDB) AcceptLease(ls *Lease, ttl time.Duration) {
	if ls.State == LeaseReserved {
		return
	}
	l.lock.Lock()
	defer l.lock.Unlock()
	for _, lease := range l.leases {
		if lease.ClientId == ls.ClientId {
			lease.State = LeaseActive
			lease.Expiry = time.Now().Add(ttl)
			return
		}
	}
}

func (l *LeaseDB) NextAvailableLease(clientId string) *Lease {
	l.lock.Lock()
	defer l.lock.Unlock()
	for _, lease := range l.leases {
		switch lease.State {
		case LeaseAvailable:
			lease.ClientId = clientId
			lease.State = LeaseOffered
			lease.Expiry = time.Now().Add(time.Second * 60)
			return lease

		case LeaseOffered:
			if strings.EqualFold(clientId, lease.ClientId) {
				lease.Expiry = time.Now().Add(time.Second * 60)
				return lease
			} else if time.Now().After(lease.Expiry) {
				lease.ClientId = clientId
				lease.Expiry = time.Now().Add(time.Second * 60)
				return lease
			}

		case LeaseActive:
			if time.Now().After(lease.Expiry) {
				lease.ClientId = clientId
				lease.State = LeaseOffered
				lease.Expiry = time.Now().Add(time.Second * 60)
				return lease
			}
		}
	}
	return nil
}

func (l *LeaseDB) ReserveLease(clientID string, reservedIP net.IP) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.reservedAddresses[clientID] = &Lease{ClientId: clientID, IP: reservedIP, State: LeaseReserved}
}

func (l *LeaseDB) PeristLeases(file string) error {
	f, err := os.OpenFile(file, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	buff := make([]byte, 1)
	var counter byte = 0
	l.lock.Lock()
	for _, lease := range l.leases {
		if lease == nil {
			break
		}
		if lease.State == LeaseActive {
			counter++
			leaseBuff := make([]byte, 0)
			leaseBuff = append(leaseBuff, byte(len(lease.ClientId)))
			leaseBuff = append(leaseBuff, []byte(lease.ClientId)...)
			leaseBuff = append(leaseBuff, byte(len(lease.Hostname)))
			leaseBuff = append(leaseBuff, []byte(lease.Hostname)...)
			leaseBuff = append(leaseBuff, lease.IP...)
			leaseBuff = append(leaseBuff, byte(lease.State))
			buff = append(buff, leaseBuff...)
		}
	}
	l.lock.Unlock()

	if counter > 0 {
		buff[0] = counter
		_, err := f.Write(buff)
		return err
	}

	return nil
}

func (l *LeaseDB) LoadLeases(file string, ttl time.Duration) error {
	f, err := os.OpenFile(file, os.O_RDONLY, 0600)
	switch {
	case err == nil:
		defer f.Close()
	case os.IsNotExist(err):
		return nil
	default:
		return err
	}

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		return nil
	}
	leaseCount := int(data[0])
	data = data[1:]
	leases := make([]*Lease, 0)
	for i := 0; i < leaseCount; i++ {
		var lease = new(Lease)
		cidLen := data[0]
		lease.ClientId = string(data[1 : 1+cidLen])
		data = data[1+cidLen:]
		hostLen := data[0]
		lease.Hostname = string(data[1 : 1+hostLen])
		data = data[1+hostLen:]
		lease.IP = data[:4]
		data = data[4:]
		lease.State = LeaseState(data[0])
		lease.Expiry = time.Now().Add(ttl)
		data = data[1:]
		leases = append(leases, lease)
	}

	l.lock.Lock()
	defer l.lock.Unlock()
	for _, lease := range leases {
		for j, createdLease := range l.leases {
			if lease.IP.Equal(createdLease.IP) {
				*l.leases[j] = *lease
				log.Debug("restore lease ", l.leases[j])
			}
		}
	}
	log.Debugf("loaded %d leases", len(leases))
	return nil
}

func (l *LeaseDB) ActiveLeases() []Lease {
	l.lock.Lock()
	defer l.lock.Unlock()
	var leases []Lease
	for i := range l.leases {
		if l.leases[i].State == LeaseActive && time.Now().Before(l.leases[i].Expiry) {
			leases = append(leases, *l.leases[i])
		}
	}
	return leases
}

func (l *LeaseDB) ReservedLeases() []Lease {
	l.lock.Lock()
	defer l.lock.Unlock()
	var leases []Lease
	for _, lease := range l.reservedAddresses {
		leases = append(leases, *lease)
	}

	return leases
}

func (l *LeaseDB) ReleaseLease(relLease *Lease) {
	l.lock.Lock()
	defer l.lock.Unlock()
	for _, lease := range l.leases {
		if strings.EqualFold(relLease.ClientId, lease.ClientId) {
			lease.Clear()
		}
	}
}
