package supporting

import (
	"encoding/json"
	"net/http"
)

type HealthStatus struct {
	Status string `json:"status"`
}

// Respond with health
func Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	health := HealthStatus{Status: "healthy"}
	json.NewEncoder(w).Encode(health)
}