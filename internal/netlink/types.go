//go:build linux
// +build linux

package netlink

import "fmt"

type Route struct {
	Family   uint8
	DstLen   uint8
	Protocol uint8
	Scope    uint8
	Dst      string
	Gateway  string
	Dev      string
	Src      string
	Metric   uint32
}

func (r *Route) String() string {
	var result string

	if r.Dst == "" && r.DstLen == 0 {
		result = "default"
	} else if r.Dst != "" {
		result = fmt.Sprintf("%s/%d", r.Dst, r.DstLen)
	}

	if r.Gateway != "" {
		result += fmt.Sprintf(" via %s", r.Gateway)
	}

	if r.Dev != "" {
		result += fmt.Sprintf(" dev %s", r.Dev)
	}

	if r.Protocol != 0 {
		protoName := protoNames[r.Protocol]
		if protoName == "" {
			protoName = fmt.Sprintf("%d", r.Protocol)
		}
		result += fmt.Sprintf(" proto %s", protoName)
	}

	if r.Scope != RT_SCOPE_UNIVERSE {
		scopeName := scopeNames[r.Scope]
		if scopeName == "" {
			scopeName = fmt.Sprintf("%d", r.Scope)
		}
		result += fmt.Sprintf(" scope %s", scopeName)
	}

	if r.Src != "" {
		result += fmt.Sprintf(" src %s", r.Src)
	}

	if r.Metric != 0 {
		result += fmt.Sprintf(" metric %d", r.Metric)
	}

	return result
}
