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

type SuspiciousTransfer struct {
    Signature string `json:"signature"`
    Amount    string `json:"amount"`
    Token     string `json:"token"`
    Sender    string `json:"sender"`
}

type DetectedTransfer struct {
	Signature string
	Amount    float64
	Sender    string
	TokenMint string // Empty if it's a SOL transfer
}

type TransferDetail struct {
	Signature       string
	Amount          float64
	Sender          string
	TokenMint       string
	IsTokenTransfer bool
}

const (
	TokenProgramID = "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"
	SystemProgramID = "11111111111111111111111111111111"
	rpcURL = "https://api.mainnet-beta.solana.com" // Replace with your preferred RPC
)

func IsValidSolanaAddress(address string) bool {
	decoded, err := base58.Decode(address)
	if err != nil {
		return false
	}
	return len(decoded) == 32
}

func GetSignatures(address string) ([]string, error) {

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

func PrintInstructionTypes(signature string) {
	reqBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "getTransaction",
		"params": []interface{}{
			signature,
			map[string]interface{}{"encoding": "jsonParsed"},
		},
	}

	jsonData, _ := json.Marshal(reqBody)
	resp, err := http.Post(rpcURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Request error:", err)
		return
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	var rpcResp map[string]interface{}
	if err := json.Unmarshal(body, &rpcResp); err != nil {
		fmt.Println("Unmarshal error:", err)
		return
	}

	result, ok := rpcResp["result"].(map[string]interface{})
	if !ok || result == nil {
		fmt.Println("No result found for transaction")
		return
	}

	transaction := result["transaction"].(map[string]interface{})
	message := transaction["message"].(map[string]interface{})
	instructions := message["instructions"].([]interface{})

	// Gather all instructions including inner ones
	meta := result["meta"].(map[string]interface{})
	if innerInstructions, ok := meta["innerInstructions"].([]interface{}); ok {
		for _, inner := range innerInstructions {
			if innerMap, ok := inner.(map[string]interface{}); ok {
				if innerInstrs, ok := innerMap["instructions"].([]interface{}); ok {
					instructions = append(instructions, innerInstrs...)
				}
			}
		}
	}

	for i, inst := range instructions {
		instMap, ok := inst.(map[string]interface{})
		if !ok {
			continue
		}
	
		parsed, hasParsed := instMap["parsed"].(map[string]interface{})
		if !hasParsed {
			continue
		}
	
		instType, hasType := parsed["type"].(string)
		if !hasType || instType != "transferChecked" {
			continue
		}
	
		info, ok := parsed["info"].(map[string]interface{})
		if !ok {
			continue
		}
	
		authority, _ := info["authority"].(string)
		mint, _ := info["mint"].(string)
	
		tokenAmountMap, ok := info["tokenAmount"].(map[string]interface{})
		if !ok {
			continue
		}
		uiAmount, _ := tokenAmountMap["uiAmount"].(float64)
	
		fmt.Printf("[%d] transferChecked Instruction:\n", i+1)
		fmt.Printf("  Authority: %s\n", authority)
		fmt.Printf("  Mint: %s\n", mint)
		fmt.Printf("  Token Amount: %.2f\n", uiAmount)
		fmt.Println("--------------------------------")
	}
	
}












