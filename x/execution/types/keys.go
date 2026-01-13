package types

const (
	// ModuleName defines the module name
	ModuleName = "execution"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for execution
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_execution"
)

// KVStore key prefixes
var (
	// ParamsKey is the key for module parameters
	ParamsKey = []byte{0x03}

	// BatchHashKeyPrefix is the prefix for batch hash storage
	// Key: 0x04 | batch_id -> txHash (32 bytes)
	BatchHashKeyPrefix = []byte{0x04}
)

// GetBatchHashKey constructs the key for batch hash storage
// Key: 0x04 | batch_id
func GetBatchHashKey(batchId string) []byte {
	batchIdBytes := []byte(batchId)
	key := make([]byte, 1+len(batchIdBytes))
	key[0] = BatchHashKeyPrefix[0]
	copy(key[1:], batchIdBytes)
	return key
}
