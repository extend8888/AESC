package types

const (
	// ModuleName defines the module name
	ModuleName = "aexburn"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

// Store key prefixes
var (
	// ParamsKey is the key for storing module params
	ParamsKey = []byte{0x01}

	// BurnStatsKey is the key for storing burn statistics
	BurnStatsKey = []byte{0x02}

	// MonthlyBurnDataPrefix is the prefix for storing monthly burn data
	MonthlyBurnDataPrefix = []byte{0x03}

	// BurnRecordPrefix is the prefix for storing burn records
	BurnRecordPrefix = []byte{0x04}
)

// GetMonthlyBurnDataKey returns the key for a specific month's burn data
func GetMonthlyBurnDataKey(monthIndex uint32) []byte {
	return append(MonthlyBurnDataPrefix, byte(monthIndex))
}

// GetBurnRecordKey returns the key for a burn record at a specific epoch
func GetBurnRecordKey(epochNumber uint64) []byte {
	bz := make([]byte, 8)
	bz[0] = byte(epochNumber >> 56)
	bz[1] = byte(epochNumber >> 48)
	bz[2] = byte(epochNumber >> 40)
	bz[3] = byte(epochNumber >> 32)
	bz[4] = byte(epochNumber >> 24)
	bz[5] = byte(epochNumber >> 16)
	bz[6] = byte(epochNumber >> 8)
	bz[7] = byte(epochNumber)
	return append(BurnRecordPrefix, bz...)
}

