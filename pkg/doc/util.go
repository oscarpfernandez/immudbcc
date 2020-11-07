package doc

import (
	"encoding/binary"
	"math"
	"regexp"
	"strconv"
	"strings"
)

func hasArrayFormat(s string) bool {
	// Checks for Arrays definitions of the format "[%d.%d]"
	return regexp.MustCompile(`^\[\d+\.\d+]$`).MatchString(s)
}

func splitArrayFormat(s string) (index int, capacity int) {
	if !hasArrayFormat(s) {
		panic("not array format")
	}

	indexCapStr := strings.Trim(s, "[]")
	valuesStr := strings.Split(indexCapStr, ".")
	index, _ = strconv.Atoi(valuesStr[0])
	capacity, _ = strconv.Atoi(valuesStr[1])

	return index, capacity
}

func PropertyNil(keys []string) PropertyEntry {
	return PropertyEntry{
		KeyURI: strings.Join(keys, "/") + "/nil",
		Value:  nil,
	}
}

func PropertyString(keys []string, value string) PropertyEntry {
	return PropertyEntry{
		KeyURI: strings.Join(keys, "/") + "/string",
		Value:  []byte(value),
	}
}

func PropertyBool(keys []string, value bool) PropertyEntry {
	return PropertyEntry{
		KeyURI: strings.Join(keys, "/") + "/bool",
		Value:  []byte(strconv.FormatBool(value)),
	}
}

func PropertyFloat64(keys []string, value float64) PropertyEntry {
	return PropertyEntry{
		KeyURI: strings.Join(keys, "/") + "/float64",
		Value:  Float64ToBinary(value),
	}
}

func Float64ToBinary(v float64) []byte {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], math.Float64bits(v))
	return buf[:]
}

func BinaryToFloat64(bytes []byte) float64 {
	bits := binary.BigEndian.Uint64(bytes)
	float := math.Float64frombits(bits)
	return float
}

func removeLastElement(s *[]string) {
	if s == nil || len(*s) == 0 {
		return
	}
	*s = (*s)[:len(*s)-1]
}
