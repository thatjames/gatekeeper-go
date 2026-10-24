package system

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/thatjames/gatekeeper-go/internal/config"
)

// useFixtures points the kernel paths at testdata and sets the config.
func useFixtures(t *testing.T, cfg *config.ConfigInstance) {
	t.Helper()
	oldMeminfo, oldRoute, oldNet, oldConfig := procMeminfo, procNetRoute, sysClassNet, config.Config
	procMeminfo = filepath.Join("testdata", "meminfo")
	procNetRoute = filepath.Join("testdata", "route")
	sysClassNet = filepath.Join("testdata", "net")
	config.Config = cfg
	t.Cleanup(func() {
		procMeminfo, procNetRoute, sysClassNet, config.Config = oldMeminfo, oldRoute, oldNet, oldConfig
	})
}

func TestWANInterfaceFromDefaultRoute(t *testing.T) {
	useFixtures(t, &config.ConfigInstance{})
	// eth0 (metric 100) beats wlan0 (600); tun0's default route is down.
	if got := wanInterface(); got != "eth0" {
		t.Errorf("expected eth0, got %q", got)
	}
}

func TestWANInterfaceConfigured(t *testing.T) {
	useFixtures(t, &config.ConfigInstance{Web: &config.Web{WANInterface: "wan"}})
	if got := wanInterface(); got != "wan" {
		t.Errorf("expected the configured wan, got %q", got)
	}
}

func TestWANInterfaceNoDefaultRoute(t *testing.T) {
	useFixtures(t, &config.ConfigInstance{})
	path := filepath.Join(t.TempDir(), "route")
	header := "Iface\tDestination\tGateway \tFlags\tRefCnt\tUse\tMetric\tMask\t\tMTU\tWindow\tIRTT\n"
	if err := os.WriteFile(path, []byte(header+"eth0\t0000000A\t00000000\t0001\t0\t0\t100\t00FFFFFF\t0\t0\t0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	procNetRoute = path
	if got := wanInterface(); got != "" {
		t.Errorf("expected no WAN interface, got %q", got)
	}
}

func TestMemoryUsage(t *testing.T) {
	useFixtures(t, &config.ConfigInstance{})
	used, total, err := memoryUsage()
	if err != nil {
		t.Fatal(err)
	}
	// 8G total, 6G available: 2G used.
	if total != 8<<30 || used != 2<<30 {
		t.Errorf("expected 2G used of 8G, got %d of %d", used, total)
	}
	if got := byteSize(used) + " / " + byteSize(total); got != "2G / 8G" {
		t.Errorf("expected \"2G / 8G\", got %q", got)
	}
}

func TestSystemInfoShowsWANAndUsedMemory(t *testing.T) {
	useFixtures(t, &config.ConfigInstance{})
	si, err := GetSystemInfo()
	if err != nil {
		t.Fatal(err)
	}
	if si["WAN Interface"] != "eth0" || si["WAN Tx"] != "1.5G" || si["WAN Rx"] != "3G" {
		t.Errorf("expected eth0 with 1.5G/3G, got %q %q %q", si["WAN Interface"], si["WAN Tx"], si["WAN Rx"])
	}
	if si["Memory"] != "2G / 8G" {
		t.Errorf("expected Memory 2G / 8G, got %q", si["Memory"])
	}
}
