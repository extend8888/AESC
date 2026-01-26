//go:build e2e_custom
// +build e2e_custom

package e2e

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
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
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/sei-protocol/x402-relayer/types"
)

// Custom ERC20 configuration - matches deployed EIP3009Token (18 decimals)
// Note: Token address is dynamically fetched from relayer's /payment-requirements endpoint
const (
	CustomTokenName    = "Test USDT"
	CustomTokenVersion = "1"
	CustomChainID      = 71603
	CustomRelayerURL   = "http://localhost:8403"
	CustomEvmRPCURL    = "http://localhost:8545"

	// Test private key (Hardhat/Anvil default account #0)
	customTestPrivateKeyHex = "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	// Relayer address
	customRelayerAddress = "0x70997970C51812dc3A010C7d01b50e0d17dc79C8"
)

// getTokenAddressFromRelayer fetches the token contract address from relayer's payment requirements
func getTokenAddressFromRelayer(t *testing.T) string {
	resp, err := http.Get(CustomRelayerURL + "/payment-requirements")
	if err != nil {
		t.Fatalf("failed to get payment requirements: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Accepts []struct {
			Asset string `json:"asset"`
		} `json:"accepts"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode payment requirements: %v", err)
	}

	if len(result.Accepts) == 0 {
		t.Fatal("no payment requirements found")
	}

	return result.Accepts[0].Asset
}

// EIP-712 type hashes
var (
	customEIP712DomainTypehash = crypto.Keccak256Hash(
		[]byte("EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)"),
	)
	customTransferWithAuthorizationTypehash = crypto.Keccak256Hash(
		[]byte("TransferWithAuthorization(address from,address to,uint256 value,uint256 validAfter,uint256 validBefore,bytes32 nonce)"),
	)
)

// computeCustomDomainSeparator computes EIP-712 domain separator for custom token
func computeCustomDomainSeparator(chainID *big.Int, tokenAddress string) [32]byte {
	nameHash := crypto.Keccak256Hash([]byte(CustomTokenName))
	versionHash := crypto.Keccak256Hash([]byte(CustomTokenVersion))

	encoded := make([]byte, 0, 160)
	encoded = append(encoded, customEIP712DomainTypehash.Bytes()...)
	encoded = append(encoded, nameHash.Bytes()...)
	encoded = append(encoded, versionHash.Bytes()...)
	encoded = append(encoded, common.LeftPadBytes(chainID.Bytes(), 32)...)
	encoded = append(encoded, common.LeftPadBytes(common.HexToAddress(tokenAddress).Bytes(), 32)...)

	return crypto.Keccak256Hash(encoded)
}

// signCustomTransferWithAuthorization signs EIP-3009 for custom token
// tokenAddress is dynamically passed to compute the correct domain separator
func signCustomTransferWithAuthorization(
	privateKey *ecdsa.PrivateKey,
	from, to common.Address,
	value *big.Int,
	validAfter, validBefore *big.Int,
	nonce [32]byte,
	chainID *big.Int,
	tokenAddress string,
) (v uint8, r, s [32]byte, err error) {
	structHash := crypto.Keccak256Hash(
		customTransferWithAuthorizationTypehash.Bytes(),
		common.LeftPadBytes(from.Bytes(), 32),
		common.LeftPadBytes(to.Bytes(), 32),
		common.LeftPadBytes(value.Bytes(), 32),
		common.LeftPadBytes(validAfter.Bytes(), 32),
		common.LeftPadBytes(validBefore.Bytes(), 32),
		nonce[:],
	)

	domainSeparator := computeCustomDomainSeparator(chainID, tokenAddress)

	digest := crypto.Keccak256Hash(
		[]byte("\x19\x01"),
		domainSeparator[:],
		structHash.Bytes(),
	)

	signature, err := crypto.Sign(digest.Bytes(), privateKey)
	if err != nil {
		return 0, [32]byte{}, [32]byte{}, err
	}

	v = signature[64] + 27
	copy(r[:], signature[:32])
	copy(s[:], signature[32:64])

	return v, r, s, nil
}

func TestCustomERC20Health(t *testing.T) {
	resp, err := http.Get(CustomRelayerURL + "/health")
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

	t.Logf("✅ Health check passed: %v", result)
}

func TestCustomERC20PaymentRequirements(t *testing.T) {
	resp, err := http.Get(CustomRelayerURL + "/payment-requirements")
	if err != nil {
		t.Fatalf("failed to get payment requirements: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	t.Logf("Payment requirements: %s", body)

	var result struct {
		Accepts []struct {
			Asset   string `json:"asset"`
			Network string `json:"network"`
			PayTo   string `json:"payTo"`
		} `json:"accepts"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}

	if len(result.Accepts) == 0 {
		t.Fatal("no payment requirements found")
	}

	// Verify the payment requirements have valid data
	accept := result.Accepts[0]
	if accept.Asset == "" {
		t.Fatal("asset address is empty")
	}
	if accept.Network != fmt.Sprintf("eip155:%d", CustomChainID) {
		t.Fatalf("unexpected network: got %s, want eip155:%d", accept.Network, CustomChainID)
	}
	if accept.PayTo == "" {
		t.Fatal("payTo address is empty")
	}

	t.Logf("✅ Payment requirements valid: asset=%s, network=%s, payTo=%s", accept.Asset, accept.Network, accept.PayTo)
}

func TestCustomERC20FullPaymentFlow(t *testing.T) {
	// Dynamically get token address from relayer
	tokenAddress := getTokenAddressFromRelayer(t)
	t.Logf("Using token address from relayer: %s", tokenAddress)

	// Load test private key
	privateKey, err := crypto.HexToECDSA(customTestPrivateKeyHex)
	if err != nil {
		t.Fatalf("failed to load private key: %v", err)
	}

	userAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
	relayer := common.HexToAddress(customRelayerAddress)

	// Create payment authorization (18 decimals: 0.01 token = 10000000000000000)
	paymentValue := big.NewInt(10000000000000000) // relay fee (0.01 token with 18 decimals)
	validAfter := big.NewInt(0)
	validBefore := big.NewInt(time.Now().Add(5 * time.Minute).Unix())
	var nonce [32]byte
	rand.Read(nonce[:])

	chainID := big.NewInt(CustomChainID)
	v, r, s, err := signCustomTransferWithAuthorization(
		privateKey, userAddress, relayer, paymentValue,
		validAfter, validBefore, nonce, chainID,
		tokenAddress, // Pass dynamic token address
	)
	if err != nil {
		t.Fatalf("failed to sign authorization: %v", err)
	}

	// Get current nonce for user address
	currentNonce := uint64(0)
	nonceResp, err := http.Post(CustomEvmRPCURL, "application/json", bytes.NewReader([]byte(fmt.Sprintf(
		`{"jsonrpc":"2.0","method":"eth_getTransactionCount","params":["%s","pending"],"id":1}`,
		userAddress.Hex(),
	))))
	if err == nil {
		defer nonceResp.Body.Close()
		var result struct {
			Result string `json:"result"`
		}
		json.NewDecoder(nonceResp.Body).Decode(&result)
		if result.Result != "" {
			nonceBig := new(big.Int)
			nonceBig.SetString(result.Result[2:], 16)
			currentNonce = nonceBig.Uint64()
		}
	}
	t.Logf("User current nonce: %d", currentNonce)

	// Create a simple signed transaction
	tx := ethtypes.NewTx(&ethtypes.LegacyTx{
		Nonce:    currentNonce,
		GasPrice: big.NewInt(1000000000),
		Gas:      21000,
		To:       &relayer,
		Value:    big.NewInt(1),
	})

	signer := ethtypes.NewEIP155Signer(chainID)
	signedTx, err := ethtypes.SignTx(tx, signer, privateKey)
	if err != nil {
		t.Fatalf("failed to sign tx: %v", err)
	}

	signedTxBytes, err := rlp.EncodeToBytes(signedTx)
	if err != nil {
		t.Fatalf("failed to encode tx: %v", err)
	}

	// Build payment payload
	payload := types.PaymentPayload{
		X402Version: 1,
		Scheme:      "exact",
		Network:     fmt.Sprintf("eip155:%d", CustomChainID),
		Payload: types.EIP3009Authorization{
			From:        userAddress,
			To:          relayer,
			Value:       paymentValue,
			ValidAfter:  validAfter,
			ValidBefore: validBefore,
			Nonce:       nonce,
			V:           v,
			R:           r,
			S:           s,
		},
	}

	payloadJSON, _ := json.Marshal(payload)
	t.Logf("Payment payload: %s", payloadJSON)

	// Encode payment as base64 (X-PAYMENT header uses base64)
	payloadBase64 := base64.StdEncoding.EncodeToString(payloadJSON)

	// Create relay request
	reqBody := struct {
		SignedTx string `json:"signedTx"`
	}{
		SignedTx: "0x" + hex.EncodeToString(signedTxBytes),
	}
	reqJSON, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", CustomRelayerURL+"/relay", bytes.NewReader(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-PAYMENT", payloadBase64)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("relay request failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	t.Logf("Relay response (status %d): %s", resp.StatusCode, body)

	// Proper assertions - test MUST fail on non-200 responses
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusPaymentRequired {
			t.Fatalf("❌ Payment rejected (402): %s", body)
		} else {
			t.Fatalf("❌ Unexpected status %d: %s", resp.StatusCode, body)
		}
	}

	// Verify response contains expected fields
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("❌ Failed to parse response JSON: %v", err)
	}

	if _, ok := result["success"]; !ok {
		t.Fatalf("❌ Response missing 'success' field")
	}
	if success, ok := result["success"].(bool); !ok || !success {
		t.Fatalf("❌ Response 'success' is not true: %v", result["success"])
	}
	if _, ok := result["txHash"]; !ok {
		t.Fatalf("❌ Response missing 'txHash' field")
	}

	t.Logf("✅ Full payment flow succeeded with custom ERC20!")
}

