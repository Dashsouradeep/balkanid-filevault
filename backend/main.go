package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/Dashsouradeep/balkanid-filevault/api"
	"github.com/gorilla/mux"
)

func main() {
	cfg := LoadConfig()
	conn := ConnectDB(cfg)
	defer conn.Close(context.Background())

	// Router
	r := mux.NewRouter()

	// User handler
	userHandler := &api.UserHandler{}
	r.HandleFunc("/users", userHandler.GetUsers).Methods("GET")

	fmt.Println("🚀 Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
