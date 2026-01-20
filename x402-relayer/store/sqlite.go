package store

import (
	"database/sql"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

// SQLiteStore implements Store interface using SQLite
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore creates a new SQLite store
func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	store := &SQLiteStore{db: db}
	if err := store.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return store, nil
}

// migrate creates the database schema
func (s *SQLiteStore) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS relay_records (
		id TEXT PRIMARY KEY,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		
		user_address TEXT NOT NULL,
		signed_tx_hash TEXT,
		
		payment_from TEXT NOT NULL,
		payment_to TEXT NOT NULL,
		payment_amount TEXT NOT NULL,
		payment_nonce TEXT NOT NULL,
		
		settle_tx_hash TEXT,
		settle_status TEXT NOT NULL DEFAULT 'pending',
		settle_error TEXT,
		settle_gas_used INTEGER DEFAULT 0,
		
		relay_tx_hash TEXT,
		relay_status TEXT NOT NULL DEFAULT 'pending',
		relay_error TEXT,
		relay_gas_used INTEGER DEFAULT 0,
		
		client_ip TEXT,
		user_agent TEXT
	);
	
	CREATE INDEX IF NOT EXISTS idx_relay_records_user_address ON relay_records(user_address);
	CREATE INDEX IF NOT EXISTS idx_relay_records_payment_from ON relay_records(payment_from);
	CREATE INDEX IF NOT EXISTS idx_relay_records_settle_status ON relay_records(settle_status);
	CREATE INDEX IF NOT EXISTS idx_relay_records_relay_status ON relay_records(relay_status);
	CREATE INDEX IF NOT EXISTS idx_relay_records_created_at ON relay_records(created_at);
	`
	_, err := s.db.Exec(schema)
	return err
}

// Create inserts a new relay record
func (s *SQLiteStore) Create(record *RelayRecord) error {
	if record.ID == "" {
		record.ID = uuid.New().String()
	}
	now := time.Now()
	record.CreatedAt = now
	record.UpdatedAt = now

	_, err := s.db.Exec(`
		INSERT INTO relay_records (
			id, created_at, updated_at,
			user_address, signed_tx_hash,
			payment_from, payment_to, payment_amount, payment_nonce,
			settle_tx_hash, settle_status, settle_error, settle_gas_used,
			relay_tx_hash, relay_status, relay_error, relay_gas_used,
			client_ip, user_agent
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		record.ID, record.CreatedAt, record.UpdatedAt,
		record.UserAddress, record.SignedTxHash,
		record.PaymentFrom, record.PaymentTo, record.PaymentAmount, record.PaymentNonce,
		record.SettleTxHash, record.SettleStatus, record.SettleError, record.SettleGasUsed,
		record.RelayTxHash, record.RelayStatus, record.RelayError, record.RelayGasUsed,
		record.ClientIP, record.UserAgent,
	)
	return err
}

// Get retrieves a record by ID
func (s *SQLiteStore) Get(id string) (*RelayRecord, error) {
	row := s.db.QueryRow(`SELECT * FROM relay_records WHERE id = ?`, id)
	return s.scanRecord(row)
}

// Update updates an existing record
func (s *SQLiteStore) Update(record *RelayRecord) error {
	record.UpdatedAt = time.Now()
	_, err := s.db.Exec(`
		UPDATE relay_records SET
			updated_at = ?,
			user_address = ?, signed_tx_hash = ?,
			payment_from = ?, payment_to = ?, payment_amount = ?, payment_nonce = ?,
			settle_tx_hash = ?, settle_status = ?, settle_error = ?, settle_gas_used = ?,
			relay_tx_hash = ?, relay_status = ?, relay_error = ?, relay_gas_used = ?,
			client_ip = ?, user_agent = ?
		WHERE id = ?
	`,
		record.UpdatedAt,
		record.UserAddress, record.SignedTxHash,
		record.PaymentFrom, record.PaymentTo, record.PaymentAmount, record.PaymentNonce,
		record.SettleTxHash, record.SettleStatus, record.SettleError, record.SettleGasUsed,
		record.RelayTxHash, record.RelayStatus, record.RelayError, record.RelayGasUsed,
		record.ClientIP, record.UserAgent,
		record.ID,
	)
	return err
}

// List returns records matching the query
func (s *SQLiteStore) List(query *RecordQuery) ([]*RelayRecord, error) {
	where, args := s.buildWhereClause(query)
	sql := fmt.Sprintf(`SELECT * FROM relay_records %s ORDER BY created_at DESC`, where)
	if query != nil && query.Limit > 0 {
		sql += fmt.Sprintf(" LIMIT %d OFFSET %d", query.Limit, query.Offset)
	}

	rows, err := s.db.Query(sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []*RelayRecord
	for rows.Next() {
		record, err := s.scanRecordRows(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, rows.Err()
}

// Count returns the number of records matching the query
func (s *SQLiteStore) Count(query *RecordQuery) (int64, error) {
	where, args := s.buildWhereClause(query)
	sql := fmt.Sprintf(`SELECT COUNT(*) FROM relay_records %s`, where)

	var count int64
	err := s.db.QueryRow(sql, args...).Scan(&count)
	return count, err
}

// Stats returns aggregated statistics
func (s *SQLiteStore) Stats(query *RecordQuery) (*RecordStats, error) {
	where, args := s.buildWhereClause(query)

	stats := &RecordStats{}

	// Total records
	countSQL := fmt.Sprintf(`SELECT COUNT(*) FROM relay_records %s`, where)
	if err := s.db.QueryRow(countSQL, args...).Scan(&stats.TotalRecords); err != nil {
		return nil, err
	}

	// Successful relays
	successSQL := fmt.Sprintf(`SELECT COUNT(*) FROM relay_records %s`, s.appendWhere(where, "relay_status = 'success'"))
	if err := s.db.QueryRow(successSQL, args...).Scan(&stats.SuccessfulRelays); err != nil {
		return nil, err
	}

	// Failed relays
	failedSQL := fmt.Sprintf(`SELECT COUNT(*) FROM relay_records %s`, s.appendWhere(where, "relay_status = 'failed'"))
	if err := s.db.QueryRow(failedSQL, args...).Scan(&stats.FailedRelays); err != nil {
		return nil, err
	}

	// Total payments and gas
	sumSQL := fmt.Sprintf(`SELECT COALESCE(SUM(CAST(payment_amount AS INTEGER)), 0), COALESCE(SUM(relay_gas_used + settle_gas_used), 0) FROM relay_records %s`, where)
	var totalPayments int64
	if err := s.db.QueryRow(sumSQL, args...).Scan(&totalPayments, &stats.TotalGasUsed); err != nil {
		return nil, err
	}
	stats.TotalPayments = big.NewInt(totalPayments).String()

	return stats, nil
}

// Close closes the database connection
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

// buildWhereClause builds WHERE clause from query
func (s *SQLiteStore) buildWhereClause(query *RecordQuery) (string, []interface{}) {
	if query == nil {
		return "", nil
	}

	var conditions []string
	var args []interface{}

	if query.UserAddress != "" {
		conditions = append(conditions, "user_address = ?")
		args = append(args, query.UserAddress)
	}
	if query.PaymentFrom != "" {
		conditions = append(conditions, "payment_from = ?")
		args = append(args, query.PaymentFrom)
	}
	if query.SettleStatus != "" {
		conditions = append(conditions, "settle_status = ?")
		args = append(args, query.SettleStatus)
	}
	if query.RelayStatus != "" {
		conditions = append(conditions, "relay_status = ?")
		args = append(args, query.RelayStatus)
	}
	if query.StartTime != nil {
		conditions = append(conditions, "created_at >= ?")
		args = append(args, *query.StartTime)
	}
	if query.EndTime != nil {
		conditions = append(conditions, "created_at <= ?")
		args = append(args, *query.EndTime)
	}

	if len(conditions) == 0 {
		return "", nil
	}
	return "WHERE " + strings.Join(conditions, " AND "), args
}

// appendWhere appends a condition to existing WHERE clause
func (s *SQLiteStore) appendWhere(where, condition string) string {
	if where == "" {
		return "WHERE " + condition
	}
	return where + " AND " + condition
}

// scanRecord scans a single row into RelayRecord
func (s *SQLiteStore) scanRecord(row *sql.Row) (*RelayRecord, error) {
	r := &RelayRecord{}
	err := row.Scan(
		&r.ID, &r.CreatedAt, &r.UpdatedAt,
		&r.UserAddress, &r.SignedTxHash,
		&r.PaymentFrom, &r.PaymentTo, &r.PaymentAmount, &r.PaymentNonce,
		&r.SettleTxHash, &r.SettleStatus, &r.SettleError, &r.SettleGasUsed,
		&r.RelayTxHash, &r.RelayStatus, &r.RelayError, &r.RelayGasUsed,
		&r.ClientIP, &r.UserAgent,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return r, err
}

// scanRecordRows scans a row from Rows into RelayRecord
func (s *SQLiteStore) scanRecordRows(rows *sql.Rows) (*RelayRecord, error) {
	r := &RelayRecord{}
	err := rows.Scan(
		&r.ID, &r.CreatedAt, &r.UpdatedAt,
		&r.UserAddress, &r.SignedTxHash,
		&r.PaymentFrom, &r.PaymentTo, &r.PaymentAmount, &r.PaymentNonce,
		&r.SettleTxHash, &r.SettleStatus, &r.SettleError, &r.SettleGasUsed,
		&r.RelayTxHash, &r.RelayStatus, &r.RelayError, &r.RelayGasUsed,
		&r.ClientIP, &r.UserAgent,
	)
	return r, err
}

