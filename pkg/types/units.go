package types

import (
	"strconv"
	"strings"
	"fmt"
)

const BVM_UNIT uint64 = 100000000

// ToAtomic: Merubah string "1.5" menjadi uint64 150000000
func (p Params) ToAtomic(amountStr string) uint64 {
	amountStr = strings.TrimSpace(amountStr)
	if amountStr == "" || amountStr == "." { return 0 }
	amountStr = strings.Replace(amountStr, ",", ".", -1)
	if strings.HasPrefix(amountStr, ".") { amountStr = "0" + amountStr }

	parts := strings.Split(amountStr, ".")
	var result uint64

	if len(parts[0]) > 0 {
		whole, _ := strconv.ParseUint(parts[0], 10, 64)
		result = whole * BVM_UNIT
	}

	if len(parts) > 1 {
		fracStr := parts[1]
		if len(fracStr) > 8 { fracStr = fracStr[:8] }
		for len(fracStr) < 8 { fracStr += "0" }
		frac, _ := strconv.ParseUint(fracStr, 10, 64)
		result += frac
	}
	return result
}

// FormatDisplay: Merubah uint64 menjadi string cantik "1.50000000"
func (p Params) FormatDisplay(amount uint64) string {
	return fmt.Sprintf("%d.%08d", amount/BVM_UNIT, amount%BVM_UNIT)
}
