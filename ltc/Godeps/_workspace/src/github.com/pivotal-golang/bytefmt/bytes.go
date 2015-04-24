// bytefmt contains helper methods and constants for converting to and from a human readable byte format.
//
//	bytefmt.ByteSize(100.5*bytefmt.MEGABYE) // "100.5M"
//	bytefmt.ByteSize(uint64(1024)) // "1K"
//
package bytefmt

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	BYTE     = 1.0
	KILOBYTE = 1024 * BYTE
	MEGABYTE = 1024 * KILOBYTE
	GIGABYTE = 1024 * MEGABYTE
	TERABYTE = 1024 * GIGABYTE
)

var bytesPattern *regexp.Regexp = regexp.MustCompile(`(?i)^(-?\d+)([KMGT]B?|B)$`)

var invalidByteQuantityError = errors.New("Byte quantity must be a positive integer with a unit of measurement like M, MB, G, or GB")

// ByteSize returns a human readable byte string, of the format 10M, 12.5K, etc.  The following units are available:
//	T Terabyte
//	G Gigabyte
//	M Megabyte
//	K Kilobyte
// the unit that would result in printing the smallest whole number is always chosen
func ByteSize(bytes uint64) string {
	unit := ""
	value := float32(bytes)

	switch {
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
		return "0"
	}

	stringValue := fmt.Sprintf("%.1f", value)
	stringValue = strings.TrimSuffix(stringValue, ".0")
	return fmt.Sprintf("%s%s", stringValue, unit)
}

// ToMegabyte parses a string formatted by ByteSize as megabytes
func ToMegabytes(s string) (uint64, error) {
	parts := bytesPattern.FindStringSubmatch(strings.TrimSpace(s))
	if len(parts) < 3 {
		return 0, invalidByteQuantityError
	}

	value, err := strconv.ParseUint(parts[1], 10, 0)
	if err != nil || value < 1 {
		return 0, invalidByteQuantityError
	}

	var bytes uint64
	unit := strings.ToUpper(parts[2])
	switch unit[:1] {
	case "T":
		bytes = value * TERABYTE
	case "G":
		bytes = value * GIGABYTE
	case "M":
		bytes = value * MEGABYTE
	case "K":
		bytes = value * KILOBYTE
	case "B":
		bytes = value * BYTE
	}

	return bytes / MEGABYTE, nil
}
