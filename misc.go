package tanklets

import (
	"fmt"
	"strings"
)

const (
	BYTE     = 1.0
	KILOBYTE = 1024 * BYTE
	MEGABYTE = 1024 * KILOBYTE
	GIGABYTE = 1024 * MEGABYTE
	TERABYTE = 1024 * GIGABYTE
)

func Bytes(b int) string {
	unit := ""
	value := float32(b)

	switch {
	case b >= TERABYTE:
		unit = "T"
		value = value / TERABYTE
	case b >= GIGABYTE:
		unit = "G"
		value = value / GIGABYTE
	case b >= MEGABYTE:
		unit = "M"
		value = value / MEGABYTE
	case b >= KILOBYTE:
		unit = "K"
		value = value / KILOBYTE
	case b >= BYTE:
		unit = "B"
	case b == 0:
		return "0"
	}

	stringValue := fmt.Sprintf("%.1f", value)
	stringValue = strings.TrimSuffix(stringValue, ".0")
	return fmt.Sprintf("%s%s", stringValue, unit)
}