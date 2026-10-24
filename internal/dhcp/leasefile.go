package dhcp

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"slices"
	"time"

	log "github.com/sirupsen/logrus"
)

// Lease file, big-endian: magic, version byte, uint32 record count, then per record
// uint16 length + client ID, uint16 length + hostname, IPv4, state byte, int64 Unix expiry (0 = none).
const (
	leaseFileMagic   = "GKLF"
	leaseFileVersion = 2
	leaseFileHeader  = len(leaseFileMagic) + 1 + 4
	// leaseRecordMin is a record with an empty client ID and hostname.
	leaseRecordMin = 2 + 2 + 4 + 1 + 8
)

var errBadLeaseFile = errors.New("bad lease file")

// SaveLeases writes the unexpired active leases and all reservations to file. The file is
// replaced by a rename, so a crash leaves either the old file or the new one.
func (l *LeasePool) SaveLeases(file string) error {
	data := l.encodeLeases(time.Now())
	dir := filepath.Dir(file)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, filepath.Base(file)+".tmp-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name()) // nothing left to remove once renamed
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), file)
}

func (l *LeasePool) encodeLeases(now time.Time) []byte {
	l.lock.Lock()
	var records []Lease
	for _, lease := range l.leases {
		if lease.State == LeaseActive && now.Before(lease.Expiry) {
			records = append(records, *lease)
		}
	}
	for _, lease := range l.reservedAddresses {
		records = append(records, *lease)
	}
	l.lock.Unlock()

	data := append([]byte(leaseFileMagic), leaseFileVersion)
	data = binary.BigEndian.AppendUint32(data, 0)
	count := uint32(0)
	for _, lease := range records {
		ip := lease.IP.To4()
		if ip == nil {
			continue
		}
		data = appendString(data, lease.ClientId)
		data = appendString(data, lease.Hostname)
		data = append(data, ip...)
		data = append(data, byte(lease.State))
		var expiry int64
		if !lease.Expiry.IsZero() {
			expiry = lease.Expiry.Unix()
		}
		data = binary.BigEndian.AppendUint64(data, uint64(expiry))
		count++
	}
	binary.BigEndian.PutUint32(data[len(leaseFileMagic)+1:], count)
	return data
}

func appendString(data []byte, s string) []byte {
	if len(s) > 0xffff {
		s = s[:0xffff]
	}
	data = binary.BigEndian.AppendUint16(data, uint16(len(s)))
	return append(data, s...)
}

// LoadLeases restores the lease file into the pool. A missing or empty file loads nothing. A file
// that does not parse is renamed <file>.bad-<UTC time>, the pool is left as it was and the error
// names the new path.
func (l *LeasePool) LoadLeases(file string) error {
	data, err := os.ReadFile(file)
	switch {
	case os.IsNotExist(err):
		return nil
	case err != nil:
		return err
	case len(data) == 0:
		return nil
	}

	saved, err := decodeLeases(data)
	if err != nil {
		bad := fmt.Sprintf("%s.bad-%s", file, time.Now().UTC().Format("20060102T150405Z"))
		if renameErr := os.Rename(file, bad); renameErr != nil {
			return fmt.Errorf("%s: %w; not moved aside: %v", file, err, renameErr)
		}
		return fmt.Errorf("%s: %w; moved to %s", file, err, bad)
	}

	now := time.Now()
	l.lock.Lock()
	defer l.lock.Unlock()
	restored := 0
	for _, s := range saved {
		if s.State == LeaseReserved {
			// The config is the only source of reservations: only the hostname comes back.
			if reserved, ok := l.reservedAddresses[s.ClientId]; ok && reserved.IP.Equal(s.IP) {
				reserved.Hostname = s.Hostname
			}
			continue
		}
		if _, reservedClient := l.reservedAddresses[s.ClientId]; reservedClient || !now.Before(s.Expiry) || l.isReserved(s.IP) {
			continue
		}
		for _, lease := range l.leases {
			if lease.IP.Equal(s.IP) {
				lease.ClientId = s.ClientId
				lease.Hostname = s.Hostname
				lease.State = LeaseActive
				lease.Expiry = s.Expiry
				log.Debug("restore lease ", lease)
				restored++
				break
			}
		}
	}
	log.Debugf("restored %d of %d saved leases", restored, len(saved))
	return nil
}

func decodeLeases(data []byte) ([]Lease, error) {
	if len(data) < leaseFileHeader || string(data[:len(leaseFileMagic)]) != leaseFileMagic {
		return nil, fmt.Errorf("%w: no %s header", errBadLeaseFile, leaseFileMagic)
	}
	if version := data[len(leaseFileMagic)]; version != leaseFileVersion {
		return nil, fmt.Errorf("%w: version %d", errBadLeaseFile, version)
	}
	count := binary.BigEndian.Uint32(data[len(leaseFileMagic)+1:])
	data = data[leaseFileHeader:]
	if uint64(count)*leaseRecordMin > uint64(len(data)) {
		return nil, fmt.Errorf("%w: %d records do not fit in %d bytes", errBadLeaseFile, count, len(data))
	}

	leases := make([]Lease, 0, count)
	for i := uint32(0); i < count; i++ {
		var lease Lease
		var ok bool
		if lease.ClientId, data, ok = readString(data); !ok {
			return nil, fmt.Errorf("%w: record %d truncated", errBadLeaseFile, i)
		}
		if lease.Hostname, data, ok = readString(data); !ok {
			return nil, fmt.Errorf("%w: record %d truncated", errBadLeaseFile, i)
		}
		if len(data) < 4+1+8 {
			return nil, fmt.Errorf("%w: record %d truncated", errBadLeaseFile, i)
		}
		lease.IP = net.IP(slices.Clone(data[:4]))
		lease.State = LeaseState(data[4])
		if lease.State != LeaseActive && lease.State != LeaseReserved {
			return nil, fmt.Errorf("%w: record %d has state %d", errBadLeaseFile, i, data[4])
		}
		if expiry := int64(binary.BigEndian.Uint64(data[5:13])); expiry != 0 {
			lease.Expiry = time.Unix(expiry, 0)
		}
		data = data[13:]
		leases = append(leases, lease)
	}
	if len(data) != 0 {
		return nil, fmt.Errorf("%w: %d trailing bytes", errBadLeaseFile, len(data))
	}
	return leases, nil
}

func readString(data []byte) (string, []byte, bool) {
	if len(data) < 2 {
		return "", nil, false
	}
	n := int(binary.BigEndian.Uint16(data))
	if len(data) < 2+n {
		return "", nil, false
	}
	return string(data[2 : 2+n]), data[2+n:], true
}

// PersistOnChange saves the pool to file after each change until stop closes: at once on the
// first change, then once more per debounce window if anything changed during it.
func (l *LeasePool) PersistOnChange(file string, debounce time.Duration, stop <-chan struct{}) {
	for {
		select {
		case <-stop:
			return
		case <-l.changed:
		}
		if err := l.SaveLeases(file); err != nil {
			log.Warn("unable to save leases: ", err.Error())
		}
		timer := time.NewTimer(debounce)
		select {
		case <-stop:
			timer.Stop()
			return
		case <-timer.C:
		}
	}
}
