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

type PoisoningResult struct {
	Count int
	Matches []struct {
		Signature      string
		FromAddress    string
		SimilarAddress string
		Amount         uint64
	}
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

func sendRPC(req RPCRequest) ([]byte, error) {
	payload, _ := json.Marshal(req)
	resp, err := http.Post(rpcURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
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

func DetectAddressPoisoning(signatures []string) (*PoisoningResult, error) {
	result := &PoisoningResult{}

	seen := make(map[string]string) // prefix => address

	for _, sig := range signatures {
		req := RPCRequest{
			Jsonrpc: "2.0",
			ID:      1,
			Method:  "getParsedTransaction",
			Params:  []interface{}{sig, map[string]interface{}{"encoding": "jsonParsed"}},
		}
		resp, err := sendRPC(req)
		if err != nil {
			continue
		}

		var txData struct {
			Result struct {
				Transaction struct {
					Message struct {
						Instructions []struct {
							Program string `json:"program"`
							Parsed  struct {
								Type string `json:"type"`
								Info map[string]interface{}
							} `json:"parsed"`
						} `json:"instructions"`
					} `json:"message"`
				} `json:"transaction"`
			} `json:"result"`
		}

		if err := json.Unmarshal(resp, &txData); err != nil {
			continue
		}

		for _, ix := range txData.Result.Transaction.Message.Instructions {
			typ := ix.Parsed.Type
			info := ix.Parsed.Info

			var from string
			var amt uint64

			if typ == "transfer" && info["source"] != nil && info["destination"] != nil {
				from = info["source"].(string)
				//to = info["destination"].(string)
				amt = uint64(info["lamports"].(float64))
			} else if typ == "transferChecked" && info["source"] != nil && info["destination"] != nil {
				from = info["source"].(string)
				//to = info["destination"].(string)
				amt = uint64(info["amount"].(float64))
			} else {
				continue
			}


			if amt > 10000 { // >0.00001 SOL or token
				continue
			}

			prefix := from[:3]
			if known, ok := seen[prefix]; ok && known != from {
				result.Count++
				result.Matches = append(result.Matches, struct {
					Signature      string
					FromAddress    string
					SimilarAddress string
					Amount         uint64
				}{
					Signature:      sig,
					FromAddress:    from,
					SimilarAddress: known,
					Amount:         amt,
				})
			}
			seen[prefix] = from
		}
	}
	return result, nil
}














