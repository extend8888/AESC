package store

import (
	"time"
)

// RelayStatus represents the status of a relay operation
type RelayStatus string

const (
	StatusPending RelayStatus = "pending"
	StatusSuccess RelayStatus = "success"
	StatusFailed  RelayStatus = "failed"
)

// RelayRecord represents a transaction relay record
type RelayRecord struct {
	// Primary key
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Request info
	UserAddress  string `json:"user_address"`   // User's EVM address
	SignedTxHash string `json:"signed_tx_hash"` // Hash of user's signed transaction

	// Payment info (EIP-3009 authorization)
	PaymentFrom   string `json:"payment_from"`   // Payer address
	PaymentTo     string `json:"payment_to"`     // Payee address (relayer)
	PaymentAmount string `json:"payment_amount"` // Payment amount in USDT smallest unit
	PaymentNonce  string `json:"payment_nonce"`  // EIP-3009 nonce (hex)

	// Settlement status (transferWithAuthorization call)
	SettleTxHash  string      `json:"settle_tx_hash"`  // Settlement transaction hash
	SettleStatus  RelayStatus `json:"settle_status"`   // pending/success/failed
	SettleError   string      `json:"settle_error"`    // Error message if failed
	SettleGasUsed uint64      `json:"settle_gas_used"` // Gas used for settlement

	// Relay status (user transaction broadcast)
	RelayTxHash  string      `json:"relay_tx_hash"`  // Relayed transaction hash
	RelayStatus  RelayStatus `json:"relay_status"`   // pending/success/failed
	RelayError   string      `json:"relay_error"`    // Error message if failed
	RelayGasUsed uint64      `json:"relay_gas_used"` // Gas used for relay

	// Metadata
	ClientIP  string `json:"client_ip"`  // Client IP address
	UserAgent string `json:"user_agent"` // Client user agent
}

// RecordQuery represents query parameters for listing records
type RecordQuery struct {
	// Filters
	UserAddress  string      `json:"user_address,omitempty"`
	PaymentFrom  string      `json:"payment_from,omitempty"`
	SettleStatus RelayStatus `json:"settle_status,omitempty"`
	RelayStatus  RelayStatus `json:"relay_status,omitempty"`

	// Time range
	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`

	// Pagination
	Limit  int `json:"limit,omitempty"`
	Offset int `json:"offset,omitempty"`
}

// RecordStats represents aggregated statistics
type RecordStats struct {
	TotalRecords     int64  `json:"total_records"`
	SuccessfulRelays int64  `json:"successful_relays"`
	FailedRelays     int64  `json:"failed_relays"`
	TotalPayments    string `json:"total_payments"` // Total USDT received
	TotalGasUsed     uint64 `json:"total_gas_used"`
}

// Store defines the interface for relay record storage
type Store interface {
	// Create a new relay record
	Create(record *RelayRecord) error

	// Get a record by ID
	Get(id string) (*RelayRecord, error)

	// Update an existing record
	Update(record *RelayRecord) error

	// List records with optional filters
	List(query *RecordQuery) ([]*RelayRecord, error)

	// Count records matching query
	Count(query *RecordQuery) (int64, error)

	// Get aggregated statistics
	Stats(query *RecordQuery) (*RecordStats, error)

	// Close the store
	Close() error
}

