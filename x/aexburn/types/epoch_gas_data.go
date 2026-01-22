package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// EpochGasData stores accumulated gas data for the current epoch
// Used to calculate the average gas usage rate across all blocks in an epoch
type EpochGasData struct {
	// TotalGasUsed is the cumulative gas used across all blocks in this epoch
	TotalGasUsed sdk.Int `json:"total_gas_used" yaml:"total_gas_used"`

	// TotalGasLimit is the cumulative gas limit across all blocks in this epoch
	TotalGasLimit sdk.Int `json:"total_gas_limit" yaml:"total_gas_limit"`

	// BlockCount is the number of blocks accumulated in this epoch
	BlockCount uint64 `json:"block_count" yaml:"block_count"`
}

// NewEpochGasData creates a new EpochGasData with zero values
func NewEpochGasData() EpochGasData {
	return EpochGasData{
		TotalGasUsed:  sdk.ZeroInt(),
		TotalGasLimit: sdk.ZeroInt(),
		BlockCount:    0,
	}
}

// Marshal encodes the EpochGasData to bytes
// Uses a simple format: blockCount (8 bytes) + gasUsedLen (4 bytes) + gasUsed + gasLimitLen (4 bytes) + gasLimit
func (e EpochGasData) Marshal() ([]byte, error) {
	gasUsedBytes := []byte(e.TotalGasUsed.String())
	gasLimitBytes := []byte(e.TotalGasLimit.String())

	// Calculate total size
	size := 8 + 4 + len(gasUsedBytes) + 4 + len(gasLimitBytes)
	buf := make([]byte, size)

	// Write block count (8 bytes, big-endian)
	buf[0] = byte(e.BlockCount >> 56)
	buf[1] = byte(e.BlockCount >> 48)
	buf[2] = byte(e.BlockCount >> 40)
	buf[3] = byte(e.BlockCount >> 32)
	buf[4] = byte(e.BlockCount >> 24)
	buf[5] = byte(e.BlockCount >> 16)
	buf[6] = byte(e.BlockCount >> 8)
	buf[7] = byte(e.BlockCount)

	// Write gas used length (4 bytes, big-endian)
	gasUsedLen := uint32(len(gasUsedBytes))
	buf[8] = byte(gasUsedLen >> 24)
	buf[9] = byte(gasUsedLen >> 16)
	buf[10] = byte(gasUsedLen >> 8)
	buf[11] = byte(gasUsedLen)

	// Write gas used bytes
	copy(buf[12:12+gasUsedLen], gasUsedBytes)

	// Write gas limit length (4 bytes, big-endian)
	offset := 12 + gasUsedLen
	gasLimitLen := uint32(len(gasLimitBytes))
	buf[offset] = byte(gasLimitLen >> 24)
	buf[offset+1] = byte(gasLimitLen >> 16)
	buf[offset+2] = byte(gasLimitLen >> 8)
	buf[offset+3] = byte(gasLimitLen)

	// Write gas limit bytes
	copy(buf[offset+4:], gasLimitBytes)

	return buf, nil
}

// Unmarshal decodes the EpochGasData from bytes
func (e *EpochGasData) Unmarshal(data []byte) error {
	if len(data) < 12 {
		*e = NewEpochGasData()
		return nil
	}

	// Read block count (8 bytes)
	e.BlockCount = uint64(data[0])<<56 | uint64(data[1])<<48 | uint64(data[2])<<40 | uint64(data[3])<<32 |
		uint64(data[4])<<24 | uint64(data[5])<<16 | uint64(data[6])<<8 | uint64(data[7])

	// Read gas used length (4 bytes)
	gasUsedLen := uint32(data[8])<<24 | uint32(data[9])<<16 | uint32(data[10])<<8 | uint32(data[11])

	if len(data) < int(12+gasUsedLen+4) {
		*e = NewEpochGasData()
		return nil
	}

	// Read gas used
	gasUsedStr := string(data[12 : 12+gasUsedLen])
	gasUsed, ok := sdk.NewIntFromString(gasUsedStr)
	if !ok {
		e.TotalGasUsed = sdk.ZeroInt()
	} else {
		e.TotalGasUsed = gasUsed
	}

	// Read gas limit length (4 bytes)
	offset := 12 + gasUsedLen
	gasLimitLen := uint32(data[offset])<<24 | uint32(data[offset+1])<<16 | uint32(data[offset+2])<<8 | uint32(data[offset+3])

	if len(data) < int(offset+4+gasLimitLen) {
		e.TotalGasLimit = sdk.ZeroInt()
		return nil
	}

	// Read gas limit
	gasLimitStr := string(data[offset+4 : offset+4+gasLimitLen])
	gasLimit, ok := sdk.NewIntFromString(gasLimitStr)
	if !ok {
		e.TotalGasLimit = sdk.ZeroInt()
	} else {
		e.TotalGasLimit = gasLimit
	}

	return nil
}

// CalculateUsageRate calculates the gas usage rate from accumulated data
// Returns a value between 0 and 1
func (e EpochGasData) CalculateUsageRate() sdk.Dec {
	if e.BlockCount == 0 || e.TotalGasLimit.IsZero() {
		return sdk.ZeroDec()
	}

	// usageRate = TotalGasUsed / TotalGasLimit
	usageRate := sdk.NewDecFromInt(e.TotalGasUsed).Quo(sdk.NewDecFromInt(e.TotalGasLimit))

	// Cap at 1.0
	if usageRate.GT(sdk.OneDec()) {
		usageRate = sdk.OneDec()
	}

	return usageRate
}

