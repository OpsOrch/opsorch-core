package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/opsorch/opsorch-core/orcherr"
)

func decodeJSON(r *http.Request, out any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(out)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, err orcherr.OpsOrchError) {
	log.Printf("API error (status=%d): code=%s, message=%s", status, err.Code, err.Message)
	writeJSON(w, status, map[string]string{"code": err.Code, "message": err.Message})
}

func asOpsOrchError(err error) *orcherr.OpsOrchError {
	var oe orcherr.OpsOrchError
	if errors.As(err, &oe) {
		return &oe
	}
	var oePtr *orcherr.OpsOrchError
	if errors.As(err, &oePtr) {
		return oePtr
	}
	return nil
}

func writeProviderError(w http.ResponseWriter, err error) {
	if oe := asOpsOrchError(err); oe != nil {
		status := http.StatusBadGateway
		switch oe.Code {
		case "not_found":
			status = http.StatusNotFound
		case "bad_request":
			status = http.StatusBadRequest
		}
		writeError(w, status, *oe)
		return
	}
	// If not an OpsOrchError, log the raw error and return a generic provider error with the actual error message
	log.Printf("Provider error (non-OpsOrchError): %v", err)
	writeError(w, http.StatusBadGateway, orcherr.OpsOrchError{Code: "provider_error", Message: err.Error()})
}
