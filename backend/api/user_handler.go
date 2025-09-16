package api

import (
	"encoding/json"
	"net/http"
)

type UserHandler struct{}

func (h *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	// For now, return dummy data
	users := []map[string]string{
		{"id": "1", "username": "admin", "email": "admin@example.com"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}
