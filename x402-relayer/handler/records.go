package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"github.com/sei-protocol/x402-relayer/store"
)

// RecordsHandler handles record query requests
type RecordsHandler struct {
	store store.Store
}

// NewRecordsHandler creates a new RecordsHandler
func NewRecordsHandler(s store.Store) *RecordsHandler {
	return &RecordsHandler{store: s}
}

// HandleList handles GET /records
func (h *RecordsHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	query := h.parseQuery(r)

	records, err := h.store.List(query)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to list records: "+err.Error())
		return
	}

	count, err := h.store.Count(query)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to count records: "+err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"records": records,
		"total":   count,
		"limit":   query.Limit,
		"offset":  query.Offset,
	})
}

// HandleGet handles GET /records/{id}
func (h *RecordsHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		h.writeError(w, http.StatusBadRequest, "id is required")
		return
	}

	record, err := h.store.Get(id)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to get record: "+err.Error())
		return
	}

	if record == nil {
		h.writeError(w, http.StatusNotFound, "record not found")
		return
	}

	h.writeJSON(w, http.StatusOK, record)
}

// HandleStats handles GET /records/stats
func (h *RecordsHandler) HandleStats(w http.ResponseWriter, r *http.Request) {
	query := h.parseQuery(r)

	stats, err := h.store.Stats(query)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "failed to get stats: "+err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, stats)
}

// parseQuery parses query parameters from request
func (h *RecordsHandler) parseQuery(r *http.Request) *store.RecordQuery {
	query := &store.RecordQuery{
		Limit:  50, // Default limit
		Offset: 0,
	}

	q := r.URL.Query()

	if userAddr := q.Get("user_address"); userAddr != "" {
		query.UserAddress = userAddr
	}

	if paymentFrom := q.Get("payment_from"); paymentFrom != "" {
		query.PaymentFrom = paymentFrom
	}

	if settleStatus := q.Get("settle_status"); settleStatus != "" {
		query.SettleStatus = store.RelayStatus(settleStatus)
	}

	if relayStatus := q.Get("relay_status"); relayStatus != "" {
		query.RelayStatus = store.RelayStatus(relayStatus)
	}

	if startTime := q.Get("start_time"); startTime != "" {
		if t, err := time.Parse(time.RFC3339, startTime); err == nil {
			query.StartTime = &t
		}
	}

	if endTime := q.Get("end_time"); endTime != "" {
		if t, err := time.Parse(time.RFC3339, endTime); err == nil {
			query.EndTime = &t
		}
	}

	if limit := q.Get("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 && l <= 100 {
			query.Limit = l
		}
	}

	if offset := q.Get("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil && o >= 0 {
			query.Offset = o
		}
	}

	return query
}

// writeJSON writes a JSON response
func (h *RecordsHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError writes an error response
func (h *RecordsHandler) writeError(w http.ResponseWriter, status int, message string) {
	h.writeJSON(w, status, map[string]string{"error": message})
}

