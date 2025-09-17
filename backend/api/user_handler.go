package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Dashsouradeep/balkanid-filevault/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserHandler struct {
	DB *pgxpool.Pool
}

func (h *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	// For now, return dummy data
	users := []map[string]string{
		{"id": "1", "username": "admin", "email": "admin@example.com"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// Register new user
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}
	user.Password = hashedPassword

	_, err = h.DB.Exec(context.Background(),
		"INSERT INTO users (username, email, password, role) VALUES ($1, $2, $3, $4)",
		user.Username, user.Email, user.Password, "user")
	if err != nil {
		http.Error(w, "Error saving user", http.StatusInternalServerError)
		fmt.Println("DB Error:", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully"})
}

// Login user
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var creds models.User
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	var dbUser models.User
	err := h.DB.QueryRow(context.Background(),
		"SELECT id, password, role FROM users WHERE username=$1 OR email=$2",
		creds.Username, creds.Email).Scan(&dbUser.ID, &dbUser.Password, &dbUser.Role)
	if err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	if !utils.CheckPassword(dbUser.Password, creds.Password) {
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		return
	}

	token, err := utils.GenerateToken(dbUser.ID, dbUser.Role)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"token": token})
}
