// +build linux
// +build arm64 amd64

package system

import (
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sys/unix"
)

func GetSystemInfo() (*SystemInfo, error) {
	var t unix.Sysinfo_t
	if err := unix.Sysinfo(&t); err != nil {
		return nil, err
	}
	hostname, _ := os.Hostname()
	return &SystemInfo{
		"Hostname": hostname,
		"Uptime":   (time.Second * time.Duration(t.Uptime)).Round(time.Second).String(),
		"Freeram":  byteSize(uint64(t.Freeram)),
		"Totalram": byteSize(uint64(t.Totalram)),
	}, nil
}

const (
	BYTE = 1 << (10 * iota)
	KILOBYTE
	MEGABYTE
	GIGABYTE
	TERABYTE
	PETABYTE
	EXABYTE
)

func byteSize(bytes uint64) string {
	unit := ""
	value := float64(bytes)

	switch {
	case bytes >= EXABYTE:
		unit = "E"
		value = value / EXABYTE
	case bytes >= PETABYTE:
		unit = "P"
		value = value / PETABYTE
	case bytes >= TERABYTE:
		unit = "T"
		value = value / TERABYTE
	case bytes >= GIGABYTE:
		unit = "G"
		value = value / GIGABYTE
	case bytes >= MEGABYTE:
		unit = "M"
		value = value / MEGABYTE
	case bytes >= KILOBYTE:
		unit = "K"
		value = value / KILOBYTE
	case bytes >= BYTE:
		unit = "B"
	case bytes == 0:
		return "0B"
	}

	result := strconv.FormatFloat(value, 'f', 1, 32)
	result = strings.TrimSuffix(result, ".0")
	return result + unit
}
