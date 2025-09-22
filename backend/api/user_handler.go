package api

import (
	"encoding/json"
	"net/http"

	"github.com/Dashsouradeep/balkanid-filevault/backend/utils"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserHandler struct {
	DB     *pgxpool.Pool
	Secret string
}

// Register - create a new user
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "❌ Invalid input", http.StatusBadRequest)
		return
	}

	// Hash password
	hashed, err := utils.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "❌ Failed to hash password", http.StatusInternalServerError)
		return
	}

	// Insert into users
	var userID int
	err = h.DB.QueryRow(r.Context(),
		`INSERT INTO users (username, email, password_hash)
		 VALUES ($1, $2, $3) RETURNING id`,
		req.Username, req.Email, hashed,
	).Scan(&userID)

	if err != nil {
		http.Error(w, "DB Error (insert user): "+err.Error(), http.StatusInternalServerError)
		return
	}

	// ✅ Initialize user_storage row immediately (100MB quota)
	_, err = h.DB.Exec(r.Context(),
		`INSERT INTO user_storage (user_id, used_bytes, quota_bytes)
		 VALUES ($1, 0, 104857600)`,
		userID,
	)
	if err != nil {
		http.Error(w, "DB Error (init storage): "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "✅ User registered successfully"})
}

// Login user
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "❌ Invalid input", http.StatusBadRequest)
		return
	}

	var id int
	var email string
	var hashed string
	err := h.DB.QueryRow(r.Context(),
		`SELECT id, email, password_hash FROM users WHERE email=$1`, req.Email).
		Scan(&id, &email, &hashed)
	if err != nil {
		http.Error(w, "❌ Invalid credentials", http.StatusUnauthorized)
		return
	}

	if !utils.CheckPasswordHash(req.Password, hashed) {
		http.Error(w, "❌ Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := utils.GenerateJWT(id, email, h.Secret)
	if err != nil {
		http.Error(w, "❌ Failed to generate token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}
