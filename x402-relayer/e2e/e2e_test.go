//go:build e2e
// +build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"
)

const (
	x402RelayerURL = "http://localhost:8402"
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

