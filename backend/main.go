package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/Dashsouradeep/balkanid-filevault/backend/api"
	"github.com/gorilla/mux"
)

func main() {
	cfg := LoadConfig()
	conn := ConnectDB(cfg)
	defer conn.Close(context.Background())

	// Router
	r := mux.NewRouter()

	// User handler (pass DB connection)
	userHandler := &api.UserHandler{DB: conn}

	// Routes
	r.HandleFunc("/users", userHandler.GetUsers).Methods("GET")
	r.HandleFunc("/register", userHandler.Register).Methods("POST")
	r.HandleFunc("/login", userHandler.Login).Methods("POST")

	fmt.Println("🚀 Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
