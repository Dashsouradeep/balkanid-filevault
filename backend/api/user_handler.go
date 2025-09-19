package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Dashsouradeep/balkanid-filevault/backend/models"
	"github.com/Dashsouradeep/balkanid-filevault/backend/utils"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserHandler handles user-related routes
type UserHandler struct {
	DB     *pgxpool.Pool
	Secret string // JWT secret
}

// Register handles new user registration
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

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "❌ Error hashing password", http.StatusInternalServerError)
		return
	}

	_, err = h.DB.Exec(r.Context(),
		"INSERT INTO users (username, email, password_hash, created_at) VALUES ($1, $2, $3, NOW())",
		req.Username, req.Email, hashedPassword,
	)
	if err != nil {
		http.Error(w, "❌ Could not create user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "✅ User registered successfully")
}

// Login handles user login and JWT issuance
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	var user models.User
	err := h.DB.QueryRow(r.Context(),
		"SELECT id, username, email, password_hash FROM users WHERE email=$1", req.Email,
	).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash)
	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Compare password
	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Generate JWT
	token, err := utils.GenerateJWT(user.ID, h.Secret)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	// Return token as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

// GetUsers fetches all users (protected route)
func (h *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query(context.Background(),
		"SELECT id, username, email, created_at FROM users")
	if err != nil {
		http.Error(w, "DB Error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.CreatedAt); err != nil {
			http.Error(w, "Row scan error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		users = append(users, u)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}
