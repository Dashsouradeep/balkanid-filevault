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
	cfg := LoadConfig()
	cfg.JWTSecret = "supersecretkey" // change later for prod
	conn := ConnectDB(cfg)
	defer conn.Close()

	// Router
	r := mux.NewRouter()

	userHandler := &api.UserHandler{DB: conn, Secret: cfg.JWTSecret}

	fileHandler := &api.FileHandler{DB: conn}

	// Create uploads dir if not exists
	os.MkdirAll("./uploads", os.ModePerm)

	r.HandleFunc("/register", userHandler.Register).Methods("POST")
	r.HandleFunc("/login", userHandler.Login).Methods("POST")
	r.HandleFunc("/users", api.AuthMiddleware(cfg.JWTSecret, userHandler.GetUsers)).Methods("GET")

	// File routes
	r.HandleFunc("/files", fileHandler.UploadFile).Methods("POST")
	r.HandleFunc("/files", fileHandler.GetFiles).Methods("GET")

	fmt.Println("🚀 Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
