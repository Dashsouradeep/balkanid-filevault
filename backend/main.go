package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Dashsouradeep/balkanid-filevault/backend/api"
	"github.com/gorilla/mux"
)

func main() {
	// Load configuration (from .env or defaults)
	cfg := LoadConfig()

	// Connect to DB
	conn := ConnectDB(cfg)
	defer conn.Close()

	// Ensure uploads folder exists
	if err := os.MkdirAll("./uploads", os.ModePerm); err != nil {
		log.Fatalf("❌ Failed to create uploads directory: %v", err)
	}

	// Router
	r := mux.NewRouter()

	// Handlers
	userHandler := &api.UserHandler{DB: conn, Secret: cfg.JWTSecret}
	fileHandler := &api.FileHandler{DB: conn, Secret: cfg.JWTSecret}
	shareHandler := &api.ShareHandler{DB: conn, Secret: cfg.JWTSecret}
	// Public routes
	r.HandleFunc("/register", userHandler.Register).Methods("POST")
	r.HandleFunc("/login", userHandler.Login).Methods("POST")

	// Protected routes
	r.HandleFunc("/users",
		api.AuthMiddleware(cfg.JWTSecret, userHandler.GetUsers),
	).Methods("GET")

	r.HandleFunc("/files",
		api.AuthMiddleware(cfg.JWTSecret, fileHandler.UploadFile),
	).Methods("POST")

	r.HandleFunc("/files",
		api.AuthMiddleware(cfg.JWTSecret, fileHandler.GetFiles),
	).Methods("GET")

	r.HandleFunc("/files/{id}",
		api.AuthMiddleware(cfg.JWTSecret, fileHandler.DownloadFile),
	).Methods("GET")

	r.HandleFunc("/share",
		api.AuthMiddleware(cfg.JWTSecret, shareHandler.ShareFile),
	).Methods("POST")

	r.HandleFunc("/shared",
		api.AuthMiddleware(cfg.JWTSecret, shareHandler.GetSharedFiles),
	).Methods("GET")

	r.HandleFunc("/files/{id}",
		api.AuthMiddleware(cfg.JWTSecret, fileHandler.DeleteFile),
	).Methods("DELETE")

	// Start server
	fmt.Println("🚀 Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
