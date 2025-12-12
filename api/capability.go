package api

import "strings"

// normalizeCapability maps plural or variant path segments to canonical capability keys.
func normalizeCapability(name string) (string, bool) {
	switch strings.ToLower(name) {
	case "incident", "incidents":
		return "incident", true
	case "alert", "alerts":
		return "alert", true
	case "log", "logs":
		return "log", true
	case "metric", "metrics":
		return "metric", true
	case "ticket", "tickets":
		return "ticket", true
	case "message", "messages", "messaging":
		return "messaging", true
	case "service", "services":
		return "service", true
	case "deployment", "deployments":
		return "deployment", true
	default:
		return "", false
	}
}
