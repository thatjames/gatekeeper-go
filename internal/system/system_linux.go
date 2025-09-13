package system

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"gitlab.com/thatjames-go/gatekeeper-go/internal/config"
	"golang.org/x/sys/unix"
)

func GetSystemInfo() (SystemInfo, error) {
	var t unix.Sysinfo_t
	if err := unix.Sysinfo(&t); err != nil {
		return nil, err
	}
	hostname, _ := os.Hostname()
	// memoryUsed := uint64(t.Totalram - t.Freeram)
	si := SystemInfo{
		"Hostname": hostname,
		"Uptime":   formatDuration((time.Second * time.Duration(t.Uptime)).Round(time.Second)),
		"Memory":   fmt.Sprintf("%s / %s", byteSize(uint64(t.Freeram)), byteSize(uint64(t.Totalram))),
	}
	if lanStats, err := getInterfaceStatsByName(config.Config.DHCP.Interface); err == nil {
		si["LAN Interface"] = config.Config.DHCP.Interface
		si["LAN Tx"] = byteSize(uint64(lanStats.TxBytes))
		si["LAN Rx"] = byteSize(uint64(lanStats.RxBytes))
	}
	if lanStats, err := getInterfaceStatsByName("ppp0"); err == nil {
		si["WAN Interface"] = "ppp0"
		si["WAN Tx"] = byteSize(uint64(lanStats.TxBytes))
		si["WAN Rx"] = byteSize(uint64(lanStats.RxBytes))
	}
	return si, nil
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
func getInterfaceStatsByName(interfaceName string) (*InterfaceStatistics, error) {
	txdat, err := ioutil.ReadFile(fmt.Sprintf("/sys/class/net/%s/statistics/tx_bytes", interfaceName))
	if err != nil {
		return nil, err
	}
	txBytes, err := strconv.ParseUint(strings.TrimRight(string(txdat), "\n"), 10, 64)
	if err != nil {
		return nil, err
	}
	rxdat, err := ioutil.ReadFile(fmt.Sprintf("/sys/class/net/%s/statistics/rx_bytes", interfaceName))
	if err != nil {
		return nil, err
	}
	rxBytes, err := strconv.ParseUint(strings.TrimRight(string(rxdat), "\n"), 10, 64)
	if err != nil {
		return nil, err
	}
	return &InterfaceStatistics{
		TxBytes: txBytes,
		RxBytes: rxBytes,
	}, nil
}
func formatDuration(d time.Duration) string {
	if d == 0 {
		return "0s"
	}
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	var parts []string
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%dd", days))
	}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}
	if minutes > 0 {
		parts = append(parts, fmt.Sprintf("%dm", minutes))
	}
	if seconds > 0 {
		parts = append(parts, fmt.Sprintf("%ds", seconds))
	}
	return strings.Join(parts, ":")
}
