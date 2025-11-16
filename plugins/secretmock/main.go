package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type rpcRequest struct {
	Method  string          `json:"method"`
	Config  map[string]any  `json:"config"`
	Payload json.RawMessage `json:"payload"`
}

type rpcResponse struct {
	Result any    `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`
}

var store = map[string]string{}

func main() {
	dec := json.NewDecoder(os.Stdin)
	enc := json.NewEncoder(os.Stdout)

	for {
		var req rpcRequest
		if err := dec.Decode(&req); err != nil {
			if err.Error() == "EOF" {
				return
			}
			_ = enc.Encode(rpcResponse{Error: err.Error()})
			return
		}

		switch req.Method {
		case "secret.put":
			var payload struct {
				Key   string `json:"key"`
				Value string `json:"value"`
			}
			if err := json.Unmarshal(req.Payload, &payload); err != nil {
				_ = enc.Encode(rpcResponse{Error: err.Error()})
				continue
			}
			store[payload.Key] = payload.Value
			_ = enc.Encode(rpcResponse{Result: map[string]string{"status": "ok"}})
		case "secret.get":
			var payload struct {
				Key string `json:"key"`
			}
			if err := json.Unmarshal(req.Payload, &payload); err != nil {
				_ = enc.Encode(rpcResponse{Error: err.Error()})
				continue
			}
			val, ok := store[payload.Key]
			if !ok {
				_ = enc.Encode(rpcResponse{Error: fmt.Sprintf("%s not found", payload.Key)})
				continue
			}
			_ = enc.Encode(rpcResponse{Result: val})
		default:
			_ = enc.Encode(rpcResponse{Error: fmt.Sprintf("unknown method %s", req.Method)})
		}
	}
}
