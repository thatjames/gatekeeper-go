package system

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/thatjames/gatekeeper-go/internal/config"
	"golang.org/x/sys/unix"
)

// Paths of the kernel files read here; tests point them at fixtures.
var (
	procMeminfo  = "/proc/meminfo"
	procNetRoute = "/proc/net/route"
	sysClassNet  = "/sys/class/net"
)

func GetSystemInfo() (SystemInfo, error) {
	var t unix.Sysinfo_t
	if err := unix.Sysinfo(&t); err != nil {
		return nil, err
	}
	hostname, _ := os.Hostname()
	used, total, err := memoryUsage()
	if err != nil {
		unit := uint64(t.Unit)
		total = uint64(t.Totalram) * unit
		used = total - uint64(t.Freeram)*unit
	}
	si := SystemInfo{
		"Hostname": hostname,
		"Uptime":   formatDuration((time.Second * time.Duration(t.Uptime)).Round(time.Second)),
		"Memory":   fmt.Sprintf("%s / %s", byteSize(used), byteSize(total)),
	}
	config.RLock()
	dhcpConfig := config.Config.DHCP
	config.RUnlock()
	if dhcpConfig != nil {
		if lanStats, err := getInterfaceStatsByName(dhcpConfig.Interface); err == nil {
			si["LAN Interface"] = dhcpConfig.Interface
			si["LAN Tx"] = byteSize(uint64(lanStats.TxBytes))
			si["LAN Rx"] = byteSize(uint64(lanStats.RxBytes))
		}
	}
	if wan := wanInterface(); wan != "" {
		if wanStats, err := getInterfaceStatsByName(wan); err == nil {
			si["WAN Interface"] = wan
			si["WAN Tx"] = byteSize(uint64(wanStats.TxBytes))
			si["WAN Rx"] = byteSize(uint64(wanStats.RxBytes))
		}
	}
	return si, nil
}

// memoryUsage reads /proc/meminfo: used is MemTotal - MemAvailable, in bytes.
func memoryUsage() (used, total uint64, err error) {
	dat, err := os.ReadFile(procMeminfo)
	if err != nil {
		return 0, 0, err
	}
	values := make(map[string]uint64)
	for _, line := range strings.Split(string(dat), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		if kb, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
			values[strings.TrimSuffix(fields[0], ":")] = kb * 1024
		}
	}
	total, okTotal := values["MemTotal"]
	available, okAvailable := values["MemAvailable"]
	if !okTotal || !okAvailable || available > total {
		return 0, 0, fmt.Errorf("%s: no MemTotal or MemAvailable", procMeminfo)
	}
	return total - available, total, nil
}

// wanInterface is Web.WANInterface when set, otherwise the interface of the
// IPv4 default route with the lowest metric, or "" when there is none.
func wanInterface() string {
	if config.Config != nil && config.Config.Web != nil && config.Config.Web.WANInterface != "" {
		return config.Config.Web.WANInterface
	}
	dat, err := os.ReadFile(procNetRoute)
	if err != nil {
		return ""
	}
	const routeUp = 0x1
	best, bestMetric := "", uint64(0)
	for _, line := range strings.Split(string(dat), "\n")[1:] {
		// Iface Destination Gateway Flags RefCnt Use Metric Mask ...
		fields := strings.Fields(line)
		if len(fields) < 8 || fields[1] != "00000000" || fields[7] != "00000000" {
			continue
		}
		flags, err1 := strconv.ParseUint(fields[3], 16, 32)
		metric, err2 := strconv.ParseUint(fields[6], 10, 32)
		if err1 != nil || err2 != nil || flags&routeUp == 0 {
			continue
		}
		if best == "" || metric < bestMetric {
			best, bestMetric = fields[0], metric
		}
	}
	return best
}

func GetNetworkInterfaces() ([]string, error) {
	interfaces := make([]string, 0)
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		if iface.Flags&net.FlagPointToPoint != 0 {
			continue
		}

		if iface.Flags&net.FlagBroadcast == 0 && iface.Flags&net.FlagMulticast == 0 {
			continue
		}

		name := strings.ToLower(iface.Name)
		if strings.HasPrefix(name, "docker") ||
			strings.HasPrefix(name, "br-") || // Docker bridges
			strings.HasPrefix(name, "bridge") ||
			strings.HasPrefix(name, "veth") || // Virtual Ethernet pairs
			strings.HasPrefix(name, "virbr") || // Libvirt bridges
			strings.HasPrefix(name, "ham") || // Hamachi bridges
			strings.Contains(name, "bridge") {
			continue
		}

		if iface.Flags&net.FlagUp != 0 && len(iface.HardwareAddr) > 0 {
			interfaces = append(interfaces, iface.Name)
		}
	}

	return interfaces, nil
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
	txdat, err := ioutil.ReadFile(fmt.Sprintf("%s/%s/statistics/tx_bytes", sysClassNet, interfaceName))
	if err != nil {
		return nil, err
	}
	txBytes, err := strconv.ParseUint(strings.TrimRight(string(txdat), "\n"), 10, 64)
	if err != nil {
		return nil, err
	}
	rxdat, err := ioutil.ReadFile(fmt.Sprintf("%s/%s/statistics/rx_bytes", sysClassNet, interfaceName))
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
