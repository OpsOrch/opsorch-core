package api

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"sync"

	"github.com/opsorch/opsorch-core/orcherr"
)

// pluginRunner executes a local plugin binary per call, passing config and payload via stdin JSON.
type pluginRunner struct {
	path   string
	config map[string]any

	mu  sync.Mutex
	cmd *exec.Cmd
	enc *json.Encoder
	dec *json.Decoder
}

func newPluginRunner(path string, config map[string]any) *pluginRunner {
	if config == nil {
		config = map[string]any{}
	}
	return &pluginRunner{path: path, config: config}
}

type rpcRequest struct {
	Method  string         `json:"method"`
	Config  map[string]any `json:"config"`
	Payload any            `json:"payload"`
}

type rpcResponse struct {
	Result json.RawMessage `json:"result,omitempty"`
	Error  *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
}

func (r *pluginRunner) call(ctx context.Context, method string, payload any, out any) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.cmd == nil {
		// Keep plugin process alive across calls; don't tie its lifetime to the request context.
		cmd := exec.CommandContext(context.Background(), r.path)
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return err
		}
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return err
		}
		if err := cmd.Start(); err != nil {
			return err
		}
		r.cmd = cmd
		r.enc = json.NewEncoder(stdin)
		r.dec = json.NewDecoder(stdout)
	}

	if err := r.enc.Encode(rpcRequest{Method: method, Config: r.config, Payload: payload}); err != nil {
		return err
	}

	var resp rpcResponse
	if err := r.dec.Decode(&resp); err != nil {
		return err
	}
	if resp.Error != nil {
		if resp.Error.Code != "" {
			return orcherr.New(resp.Error.Code, resp.Error.Message, nil)
		}
		return fmt.Errorf(resp.Error.Message)
	}
	if out != nil && resp.Result != nil {
		if err := json.Unmarshal(resp.Result, out); err != nil {
			return err
		}
	}
	return nil
}
