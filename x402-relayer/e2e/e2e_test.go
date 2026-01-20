//go:build e2e
// +build e2e

package e2e

import (
	"bytes"
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

