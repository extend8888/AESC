package x402relayer

import (
	"context"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/sei-protocol/x402-relayer/config"
	"github.com/sei-protocol/x402-relayer/facilitator"
	"github.com/sei-protocol/x402-relayer/handler"
	"github.com/sei-protocol/x402-relayer/middleware"
	"github.com/sei-protocol/x402-relayer/relayer"
	"github.com/sei-protocol/x402-relayer/store"
)

// Server represents the x402 HTTP server
type Server struct {
	config     *config.Config
	httpServer *http.Server
	router     *mux.Router

	// Components
	verifier       *facilitator.Verifier
	balanceChecker *facilitator.BalanceChecker
	settler        *facilitator.Settler
	broadcaster    *relayer.Broadcaster
	gasEstimator   *relayer.GasEstimator
	store          store.Store
}

// NewServer creates a new x402 server
func NewServer(cfg *config.Config) (*Server, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Parse chain ID from network ID (e.g., "eip155:1" -> 1)
	chainID, err := parseChainID(cfg.NetworkID)
	if err != nil {
		return nil, fmt.Errorf("invalid network ID: %w", err)
	}

	// Parse relay fee
	relayFee, ok := new(big.Int).SetString(cfg.RelayFeePerTx, 10)
	if !ok {
		return nil, fmt.Errorf("invalid relay fee: %s", cfg.RelayFeePerTx)
	}

	// Get private key
	privateKey := cfg.GetPrivateKey()
	if privateKey == "" {
		return nil, fmt.Errorf("private key is required")
	}

	// Initialize components
	verifier := facilitator.NewVerifier(chainID)

	balanceChecker, err := facilitator.NewBalanceChecker(cfg.EVMRPC)
	if err != nil {
		return nil, fmt.Errorf("failed to create balance checker: %w", err)
	}

	settler, err := facilitator.NewSettler(cfg.EVMRPC, privateKey, chainID)
	if err != nil {
		balanceChecker.Close()
		return nil, fmt.Errorf("failed to create settler: %w", err)
	}

	broadcaster, err := relayer.NewBroadcaster(cfg.EVMRPC, chainID)
	if err != nil {
		balanceChecker.Close()
		settler.Close()
		return nil, fmt.Errorf("failed to create broadcaster: %w", err)
	}

	gasEstimator, err := relayer.NewGasEstimator(cfg.EVMRPC, relayFee)
	if err != nil {
		balanceChecker.Close()
		settler.Close()
		broadcaster.Close()
		return nil, fmt.Errorf("failed to create gas estimator: %w", err)
	}

	// Initialize store
	dbStore, err := store.NewSQLiteStore(cfg.DBPath)
	if err != nil {
		balanceChecker.Close()
		settler.Close()
		broadcaster.Close()
		gasEstimator.Close()
		return nil, fmt.Errorf("failed to create store: %w", err)
	}

	// Create router
	router := mux.NewRouter()

	server := &Server{
		config:         cfg,
		router:         router,
		verifier:       verifier,
		balanceChecker: balanceChecker,
		settler:        settler,
		broadcaster:    broadcaster,
		gasEstimator:   gasEstimator,
		store:          dbStore,
	}

	// Setup routes
	server.setupRoutes()

	return server, nil
}

// setupRoutes configures the HTTP routes
func (s *Server) setupRoutes() {
	// Create handlers
	relayHandler := handler.NewRelayHandler(
		s.config,
		s.verifier,
		s.balanceChecker,
		s.settler,
		s.broadcaster,
		s.gasEstimator,
		s.store,
	)

	// Create records handler
	recordsHandler := handler.NewRecordsHandler(s.store)

	// Create payment middleware
	paymentMiddleware := middleware.NewPaymentMiddleware(
		s.config,
		s.verifier,
		s.balanceChecker,
	)

	// Health check (no payment required)
	s.router.HandleFunc("/health", handler.HealthHandler).Methods("GET")

	// Relay endpoint (payment required) - wrap handler with middleware
	relayWithPayment := paymentMiddleware.Middleware(http.HandlerFunc(relayHandler.HandleRelay))
	s.router.Handle("/relay", relayWithPayment).Methods("POST")

	// Payment requirements endpoint
	s.router.HandleFunc("/payment-requirements", relayHandler.HandlePaymentRequirements).Methods("GET")

	// Records endpoints (for debugging/admin)
	// Note: /records/stats must be registered before /records/{id} to avoid being matched as an ID
	s.router.HandleFunc("/records/stats", recordsHandler.HandleStats).Methods("GET")
	s.router.HandleFunc("/records", recordsHandler.HandleList).Methods("GET")
	s.router.HandleFunc("/records/{id}", recordsHandler.HandleGet).Methods("GET")
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.config.Port),
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s.httpServer.ListenAndServe()
}

// Stop gracefully stops the HTTP server
func (s *Server) Stop(ctx context.Context) error {
	// Close components
	s.balanceChecker.Close()
	s.settler.Close()
	s.broadcaster.Close()
	s.gasEstimator.Close()
	s.store.Close()

	// Shutdown HTTP server
	return s.httpServer.Shutdown(ctx)
}

// parseChainID parses chain ID from CAIP-2 network ID (e.g., "eip155:1" -> 1)
func parseChainID(networkID string) (*big.Int, error) {
	var chainIDStr string
	_, err := fmt.Sscanf(networkID, "eip155:%s", &chainIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid CAIP-2 network ID format: %s", networkID)
	}

	chainID, ok := new(big.Int).SetString(chainIDStr, 10)
	if !ok {
		return nil, fmt.Errorf("invalid chain ID: %s", chainIDStr)
	}

	return chainID, nil
}

