package util

import (
	"strconv"
	"strings"
)

const (
	BYTE = 1 << (10 * iota)
	KILOBYTE
	MEGABYTE
	GIGABYTE
	// TERABYTE
	// PETABYTE
	// EXABYTE
)

func ByteSize(bytes int) string {
	unit := ""
	value := float64(bytes)

	switch {
	// case bytes >= EXABYTE:
	// 	unit = "E"
	// 	value = value / EXABYTE
	// case bytes >= PETABYTE:
	// 	unit = "P"
	// 	value = value / PETABYTE
	// case bytes >= TERABYTE:
	// 	unit = "T"
	// 	value = value / TERABYTE
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

	result := strconv.FormatFloat(value, 'f', 1, 64)
	result = strings.TrimSuffix(result, ".0")
	return result + unit
}
