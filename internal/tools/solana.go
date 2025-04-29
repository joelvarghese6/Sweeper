package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/mr-tron/base58"
)

type RPCRequest struct {
	Jsonrpc string        `json:"jsonrpc"`
	ID      int           `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}


type TransferInfo struct {
	OtherAccount     string  // The counterparty account
	Amount           float64 // In SOL
	Signature        string  // The tx signature
	Dust             bool    // true if amount <= 0.000001
	MultipleAccounts bool    // true if tx has multiple system_transfer instructions
}

func IsValidSolanaAddress(address string) bool {
	decoded, err := base58.Decode(address)
	if err != nil {
		return false
	}
	return len(decoded) == 32
}

func GetSignatures(address string) ([]string, error) {
	rpcURL := "https://mainnet.helius-rpc.com/?api-key=0f31c860-68c3-4d89-bc63-a2f8957a0603"

	// Prepare JSON-RPC request
	reqBody := RPCRequest{
		Jsonrpc: "2.0",
		ID:      1,
		Method:  "getSignaturesForAddress",
		Params:  []interface{}{address, map[string]int{"limit": 30}}, // adjust limit as needed
	}

	jsonData, _ := json.Marshal(reqBody)
	resp, err := http.Post(rpcURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	body, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(body, &result)

	// Extract signatures
	var signatures []string
	for _, entry := range result["result"].([]interface{}) {
		sig := entry.(map[string]interface{})["signature"].(string)
		signatures = append(signatures, sig)
	}

	return signatures, nil
}

func GetTransaction(signature string) (map[string]interface{}, error) {
	rpcURL := "https://mainnet.helius-rpc.com/?api-key=0f31c860-68c3-4d89-bc63-a2f8957a0603"

	reqBody := RPCRequest{
		Jsonrpc: "2.0",
		ID:      1,
		Method:  "getTransaction",
		Params:  []interface{}{signature, "json"},
	}

	jsonData, _ := json.Marshal(reqBody)
	resp, err := http.Post(rpcURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var rpcResp map[string]interface{}
	body, _ := ioutil.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &rpcResp); err != nil {
		return nil, err
	}

	result, ok := rpcResp["result"].(map[string]interface{})
	if !ok || result == nil {
		return nil, fmt.Errorf("no transaction found for signature: %s", signature)
	}

	return result, nil
}

func AnalyzeSystemTransferToAddress(signature string, address string) ([]TransferInfo, error) {
	rpcURL := "https://mainnet.helius-rpc.com/?api-key=0f31c860-68c3-4d89-bc63-a2f8957a0603"

	reqBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "getTransaction",
		"params":  []interface{}{signature, "json"},
	}

	jsonData, _ := json.Marshal(reqBody)
	resp, err := http.Post(rpcURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var rpcResp map[string]interface{}
	body, _ := ioutil.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &rpcResp); err != nil {
		return nil, err
	}

	result, ok := rpcResp["result"].(map[string]interface{})
	if !ok || result == nil {
		return nil, fmt.Errorf("invalid transaction or result missing")
	}

	tx, _ := result["transaction"].(map[string]interface{})
	message, _ := tx["message"].(map[string]interface{})
	accountKeys, _ := message["accountKeys"].([]interface{})
	instructions, _ := message["instructions"].([]interface{})

	meta, _ := result["meta"].(map[string]interface{})
	preBalances, _ := meta["preBalances"].([]interface{})
	postBalances, _ := meta["postBalances"].([]interface{})

	var systemTransferCount int

	// First pass: count system_transfer instructions
	for _, instr := range instructions {
		inst := instr.(map[string]interface{})
		programIndex := int(inst["programIdIndex"].(float64))
		program := accountKeys[programIndex].(string)
		if program == "11111111111111111111111111111111" {
			systemTransferCount++
		}
	}

	var infos []TransferInfo

	// Second pass: extract matching instructions
	for _, instr := range instructions {
		inst := instr.(map[string]interface{})
		programIndex := int(inst["programIdIndex"].(float64))
		program := accountKeys[programIndex].(string)

		if program != "11111111111111111111111111111111" {
			continue // skip non-system transfers
		}

		accountIndices := inst["accounts"].([]interface{})
		if len(accountIndices) < 2 {
			continue
		}

		fromIdx := int(accountIndices[0].(float64))
		toIdx := int(accountIndices[1].(float64))
		fromAddr := accountKeys[fromIdx].(string)
		toAddr := accountKeys[toIdx].(string)

		if fromAddr != address && toAddr != address {
			continue
		}

		//fromPre := preBalances[fromIdx].(float64)
		//fromPost := postBalances[fromIdx].(float64)
		toPre := preBalances[toIdx].(float64)
		toPost := postBalances[toIdx].(float64)

		amount := toPost - toPre
		if amount <= 0 {
			continue
		}
		amountSOL := amount / 1e9

		// Set counterparty
		other := fromAddr
		if fromAddr == address {
			other = toAddr
		}

		infos = append(infos, TransferInfo{
			OtherAccount:     other,
			Amount:           amountSOL,
			Signature:        signature,
			Dust:             amountSOL <= 0.000001,
			MultipleAccounts: systemTransferCount > 1,
		})
	}

	return infos, nil
}











