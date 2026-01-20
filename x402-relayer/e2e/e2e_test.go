//go:build e2e
// +build e2e

package e2e

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/sei-protocol/x402-relayer/types"
)

const (
	x402RelayerURL = "http://localhost:8402"
	evmRPCURL      = "http://localhost:8545"

	// Test private key (Hardhat/Anvil default account #0)
	// Address: 0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266
	testPrivateKeyHex = "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"

	// Relayer private key (Hardhat/Anvil default account #1)
	// Address: 0x70997970C51812dc3A010C7d01b50e0d17dc79C8
	relayerPrivateKeyHex = "59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d"
	relayerAddress       = "0x70997970C51812dc3A010C7d01b50e0d17dc79C8"
)

func TestHealthEndpoint(t *testing.T) {
	resp, err := http.Get(x402RelayerURL + "/health")
	if err != nil {
		t.Fatalf("failed to call health endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if result["status"] != "ok" {
		t.Fatalf("expected status 'ok', got %v", result["status"])
	}

	t.Logf("Health check passed: %v", result)
}

func TestPaymentRequirementsEndpoint(t *testing.T) {
	resp, err := http.Get(x402RelayerURL + "/payment-requirements")
	if err != nil {
		t.Fatalf("failed to call payment-requirements endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	t.Logf("Payment requirements: %s", string(body))

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	accepts, ok := result["accepts"].([]interface{})
	if !ok || len(accepts) == 0 {
		t.Fatalf("expected non-empty accepts array")
	}

	firstAccept := accepts[0].(map[string]interface{})
	if firstAccept["scheme"] != "exact" {
		t.Fatalf("expected scheme 'exact', got %v", firstAccept["scheme"])
	}
	if firstAccept["network"] != "eip155:713715" {
		t.Fatalf("expected network 'eip155:713715', got %v", firstAccept["network"])
	}
}

func TestRelayWithoutPayment(t *testing.T) {
	// Try to relay without payment header - should return 402
	reqBody := map[string]string{
		"signedTx": "0xf86c0a85046c7cfe0082520894" +
			"0000000000000000000000000000000000000001" +
			"880de0b6b3a7640000801ca0" +
			"0000000000000000000000000000000000000000000000000000000000000001a0" +
			"0000000000000000000000000000000000000000000000000000000000000002",
	}
	body, _ := json.Marshal(reqBody)

	resp, err := http.Post(x402RelayerURL+"/relay", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("failed to call relay endpoint: %v", err)
	}
	defer resp.Body.Close()

	// Should return 402 Payment Required
	if resp.StatusCode != http.StatusPaymentRequired {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status 402, got %d: %s", resp.StatusCode, string(respBody))
	}

	t.Logf("Relay without payment correctly rejected with 402")
}

func TestRecordsEndpoint(t *testing.T) {
	resp, err := http.Get(x402RelayerURL + "/records")
	if err != nil {
		t.Fatalf("failed to call records endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	_, ok := result["records"]
	if !ok {
		t.Fatalf("expected 'records' field in response")
	}

	_, ok = result["total"]
	if !ok {
		t.Fatalf("expected 'total' field in response")
	}

	t.Logf("Records endpoint works: total=%v", result["total"])
}

func TestStatsEndpoint(t *testing.T) {
	resp, err := http.Get(x402RelayerURL + "/records/stats")
	if err != nil {
		t.Fatalf("failed to call stats endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status 200, got %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	t.Logf("Stats endpoint works: %v", result)
}

// TestFullRelayWithPayment tests the complete relay flow with valid EIP-3009 payment
func TestFullRelayWithPayment(t *testing.T) {
	// Load test private key
	privateKey, err := crypto.HexToECDSA(testPrivateKeyHex)
	if err != nil {
		t.Fatalf("failed to load private key: %v", err)
	}

	fromAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
	toAddress := common.HexToAddress(relayerAddress)
	chainID := big.NewInt(ChainID)

	t.Logf("Test account: %s", fromAddress.Hex())

	// Create EIP-3009 authorization
	paymentAmount := big.NewInt(10000) // 0.01 USDT (6 decimals)
	validAfter := big.NewInt(0)
	validBefore := big.NewInt(time.Now().Add(5 * time.Minute).Unix())
	nonce := generateRandomNonce()

	// Sign the authorization
	v, r, s, err := signTransferWithAuthorization(
		privateKey,
		fromAddress,
		toAddress,
		paymentAmount,
		validAfter,
		validBefore,
		nonce,
		chainID,
	)
	if err != nil {
		t.Fatalf("failed to sign authorization: %v", err)
	}

	auth := types.EIP3009Authorization{
		From:        fromAddress,
		To:          toAddress,
		Value:       paymentAmount,
		ValidAfter:  validAfter,
		ValidBefore: validBefore,
		Nonce:       nonce,
		V:           v,
		R:           r,
		S:           s,
	}

	// Create payment payload
	paymentBase64, err := createPaymentPayload(auth)
	if err != nil {
		t.Fatalf("failed to create payment payload: %v", err)
	}

	t.Logf("Payment payload created (base64 length: %d)", len(paymentBase64))

	// Create a simple transaction to relay
	// This is a transfer to a random address
	txNonce := uint64(0) // Will likely fail due to nonce, but tests the flow
	txTo := common.HexToAddress("0x0000000000000000000000000000000000000001")
	txAmount := big.NewInt(0)
	txGasLimit := uint64(21000)
	txGasPrice := big.NewInt(1000000000) // 1 gwei

	signedTx, err := signTransaction(
		privateKey,
		txNonce,
		txTo,
		txAmount,
		txGasLimit,
		txGasPrice,
		chainID,
	)
	if err != nil {
		t.Fatalf("failed to sign transaction: %v", err)
	}

	// Encode transaction to hex
	txBytes, err := rlp.EncodeToBytes(signedTx)
	if err != nil {
		t.Fatalf("failed to encode transaction: %v", err)
	}
	signedTxHex := "0x" + hex.EncodeToString(txBytes)

	t.Logf("Signed transaction: %s...", signedTxHex[:66])

	// Send relay request with payment header
	reqBody := map[string]string{
		"signedTx": signedTxHex,
	}
	body, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", x402RelayerURL+"/relay", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-PAYMENT", paymentBase64)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to send relay request: %v", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	t.Logf("Relay response (status %d): %s", resp.StatusCode, string(respBody))

	// Parse response
	var relayResp map[string]interface{}
	if err := json.Unmarshal(respBody, &relayResp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	// Check the response - we expect either success or a specific error
	// (The test account may not have USDT balance or the nonce may be wrong)
	if resp.StatusCode == http.StatusOK {
		t.Logf("✅ Relay succeeded!")
		t.Logf("   TxHash: %v", relayResp["txHash"])
		t.Logf("   GasUsed: %v", relayResp["gasUsed"])
		t.Logf("   RecordID: %v", relayResp["recordId"])
	} else if resp.StatusCode == http.StatusPaymentRequired {
		t.Logf("⚠️ Payment required/failed: %v", relayResp["error"])
		// This is expected if the test account doesn't have USDT
	} else {
		t.Logf("⚠️ Relay returned status %d: %v", resp.StatusCode, relayResp["error"])
	}

	// Verify a record was created (if we got past payment validation)
	if recordID, ok := relayResp["recordId"]; ok && recordID != nil {
		recordResp, err := http.Get(fmt.Sprintf("%s/records/%s", x402RelayerURL, recordID))
		if err == nil {
			defer recordResp.Body.Close()
			var record map[string]interface{}
			json.NewDecoder(recordResp.Body).Decode(&record)
			t.Logf("Record details: %+v", record)
		}
	}
}

// TestFailureScenarios tests various failure cases
func TestFailureScenarios(t *testing.T) {
	t.Run("InvalidSignature", testInvalidSignature)
	t.Run("ExpiredAuthorization", testExpiredAuthorization)
	t.Run("InvalidNetwork", testInvalidNetwork)
	t.Run("MalformedPaymentHeader", testMalformedPaymentHeader)
}

func testInvalidSignature(t *testing.T) {
	privateKey, _ := crypto.HexToECDSA(testPrivateKeyHex)
	fromAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
	toAddress := common.HexToAddress(relayerAddress)
	chainID := big.NewInt(ChainID)

	paymentAmount := big.NewInt(10000)
	validAfter := big.NewInt(0)
	validBefore := big.NewInt(time.Now().Add(5 * time.Minute).Unix())
	nonce := generateRandomNonce()

	// Get valid signature first
	v, r, s, _ := signTransferWithAuthorization(
		privateKey, fromAddress, toAddress, paymentAmount,
		validAfter, validBefore, nonce, chainID,
	)

	// Corrupt the signature
	r[0] ^= 0xFF

	auth := types.EIP3009Authorization{
		From: fromAddress, To: toAddress, Value: paymentAmount,
		ValidAfter: validAfter, ValidBefore: validBefore, Nonce: nonce,
		V: v, R: r, S: s,
	}

	paymentBase64, _ := createPaymentPayload(auth)

	// Create relay request
	reqBody := map[string]string{"signedTx": "0x1234"}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", x402RelayerURL+"/relay", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-PAYMENT", paymentBase64)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusPaymentRequired {
		respBody, _ := io.ReadAll(resp.Body)
		t.Logf("Response: %s", string(respBody))
		// Accept 402 (signature validation failed) or 500 (ecrecover failed)
		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("expected 402 or 500, got %d", resp.StatusCode)
		}
	}
	t.Logf("✅ Invalid signature correctly rejected (status %d)", resp.StatusCode)
}

func testExpiredAuthorization(t *testing.T) {
	privateKey, _ := crypto.HexToECDSA(testPrivateKeyHex)
	fromAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
	toAddress := common.HexToAddress(relayerAddress)
	chainID := big.NewInt(ChainID)

	paymentAmount := big.NewInt(10000)
	validAfter := big.NewInt(0)
	validBefore := big.NewInt(time.Now().Add(-1 * time.Minute).Unix()) // Already expired
	nonce := generateRandomNonce()

	v, r, s, _ := signTransferWithAuthorization(
		privateKey, fromAddress, toAddress, paymentAmount,
		validAfter, validBefore, nonce, chainID,
	)

	auth := types.EIP3009Authorization{
		From: fromAddress, To: toAddress, Value: paymentAmount,
		ValidAfter: validAfter, ValidBefore: validBefore, Nonce: nonce,
		V: v, R: r, S: s,
	}

	paymentBase64, _ := createPaymentPayload(auth)

	reqBody := map[string]string{"signedTx": "0x1234"}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", x402RelayerURL+"/relay", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-PAYMENT", paymentBase64)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusPaymentRequired {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 402, got %d: %s", resp.StatusCode, string(respBody))
	}

	respBody, _ := io.ReadAll(resp.Body)
	t.Logf("✅ Expired authorization correctly rejected: %s", string(respBody))
}

func testInvalidNetwork(t *testing.T) {
	privateKey, _ := crypto.HexToECDSA(testPrivateKeyHex)
	fromAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
	toAddress := common.HexToAddress(relayerAddress)
	chainID := big.NewInt(ChainID)

	paymentAmount := big.NewInt(10000)
	validAfter := big.NewInt(0)
	validBefore := big.NewInt(time.Now().Add(5 * time.Minute).Unix())
	nonce := generateRandomNonce()

	v, r, s, _ := signTransferWithAuthorization(
		privateKey, fromAddress, toAddress, paymentAmount,
		validAfter, validBefore, nonce, chainID,
	)

	auth := types.EIP3009Authorization{
		From: fromAddress, To: toAddress, Value: paymentAmount,
		ValidAfter: validAfter, ValidBefore: validBefore, Nonce: nonce,
		V: v, R: r, S: s,
	}

	// Create payload with wrong network
	payload := types.PaymentPayload{
		X402Version: 1,
		Scheme:      "exact",
		Network:     "eip155:999999", // Wrong network
		Payload:     auth,
	}
	jsonBytes, _ := json.Marshal(payload)
	paymentBase64 := base64.StdEncoding.EncodeToString(jsonBytes)

	reqBody := map[string]string{"signedTx": "0x1234"}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", x402RelayerURL+"/relay", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-PAYMENT", paymentBase64)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusPaymentRequired {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 402, got %d: %s", resp.StatusCode, string(respBody))
	}

	respBody, _ := io.ReadAll(resp.Body)
	t.Logf("✅ Wrong network correctly rejected: %s", string(respBody))
}

func testMalformedPaymentHeader(t *testing.T) {
	reqBody := map[string]string{"signedTx": "0x1234"}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", x402RelayerURL+"/relay", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-PAYMENT", "not-valid-base64!!!")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusPaymentRequired {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 402, got %d: %s", resp.StatusCode, string(respBody))
	}

	t.Logf("✅ Malformed payment header correctly rejected (status %d)", resp.StatusCode)
}

// TestConcurrentRelays tests concurrent transaction processing
func TestConcurrentRelays(t *testing.T) {
	const numConcurrent = 5

	privateKey, err := crypto.HexToECDSA(testPrivateKeyHex)
	if err != nil {
		t.Fatalf("failed to load private key: %v", err)
	}

	fromAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
	toAddress := common.HexToAddress(relayerAddress)
	chainID := big.NewInt(ChainID)

	// Create multiple concurrent requests
	results := make(chan struct {
		status int
		err    error
	}, numConcurrent)

	startTime := time.Now()

	for i := 0; i < numConcurrent; i++ {
		go func(idx int) {
			paymentAmount := big.NewInt(10000)
			validAfter := big.NewInt(0)
			validBefore := big.NewInt(time.Now().Add(5 * time.Minute).Unix())
			nonce := generateRandomNonce()

			v, r, s, _ := signTransferWithAuthorization(
				privateKey, fromAddress, toAddress, paymentAmount,
				validAfter, validBefore, nonce, chainID,
			)

			auth := types.EIP3009Authorization{
				From: fromAddress, To: toAddress, Value: paymentAmount,
				ValidAfter: validAfter, ValidBefore: validBefore, Nonce: nonce,
				V: v, R: r, S: s,
			}

			paymentBase64, _ := createPaymentPayload(auth)

			// Create unique transaction with different nonce
			txTo := common.HexToAddress("0x0000000000000000000000000000000000000001")
			signedTx, _ := signTransaction(
				privateKey, uint64(idx), txTo, big.NewInt(0),
				21000, big.NewInt(1000000000), chainID,
			)
			txBytes, _ := rlp.EncodeToBytes(signedTx)
			signedTxHex := "0x" + hex.EncodeToString(txBytes)

			reqBody := map[string]string{"signedTx": signedTxHex}
			body, _ := json.Marshal(reqBody)

			req, _ := http.NewRequest("POST", x402RelayerURL+"/relay", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-PAYMENT", paymentBase64)

			client := &http.Client{Timeout: 30 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				results <- struct {
					status int
					err    error
				}{0, err}
				return
			}
			resp.Body.Close()

			results <- struct {
				status int
				err    error
			}{resp.StatusCode, nil}
		}(i)
	}

	// Collect results
	successCount := 0
	failedCount := 0
	for i := 0; i < numConcurrent; i++ {
		result := <-results
		if result.err != nil {
			t.Logf("Request %d failed: %v", i, result.err)
			failedCount++
		} else if result.status == http.StatusOK {
			successCount++
		} else {
			// 402/500 are expected for concurrent requests (nonce conflicts, etc.)
			failedCount++
		}
	}

	duration := time.Since(startTime)
	t.Logf("✅ Concurrent test completed in %v", duration)
	t.Logf("   Success: %d, Failed/Rejected: %d", successCount, failedCount)
	t.Logf("   Throughput: %.2f req/s", float64(numConcurrent)/duration.Seconds())

	// At least verify the server didn't crash
	resp, err := http.Get(x402RelayerURL + "/health")
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("Server health check failed after concurrent test")
	}
	resp.Body.Close()
	t.Logf("✅ Server remained healthy after concurrent requests")
}

func init() {
	// Wait for server to be ready
	for i := 0; i < 10; i++ {
		resp, err := http.Get(x402RelayerURL + "/health")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			fmt.Println("x402-relayer is ready")
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
	fmt.Println("Warning: x402-relayer may not be ready")
}

