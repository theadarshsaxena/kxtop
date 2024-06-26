package analytics

import (
	"strconv"
	"strings"
)

func ConvertToMi(memory string) int64 {
	if strings.HasSuffix(memory, "Ki") {
		memoryInt, _ := strconv.ParseInt(strings.TrimSuffix(memory, "Ki"), 10, 64)
		return memoryInt / 1024
	}
	if strings.HasSuffix(memory, "Mi") {
		memoryInt, _ := strconv.ParseInt(strings.TrimSuffix(memory, "Mi"), 10, 64)
		return memoryInt
	}
	if strings.HasSuffix(memory, "Gi") {
		memoryInt, _ := strconv.ParseInt(strings.TrimSuffix(memory, "Gi"), 10, 64)
		return memoryInt * 1024
	}
	return 0
}
