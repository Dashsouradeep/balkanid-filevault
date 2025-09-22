package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"github.com/Dashsouradeep/balkanid-filevault/backend/api"
	"github.com/Dashsouradeep/balkanid-filevault/backend/db"
)

func main() {
	// Load DB
	pool, err := db.ConnectDB()
	if err != nil {
		log.Fatal("❌ Failed to connect DB: ", err)
	}
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "supersecret" // fallback for dev
	}

	// Handlers
	userHandler := &api.UserHandler{DB: pool, Secret: secret}
	fileHandler := &api.FileHandler{DB: pool, Secret: secret}
	shareHandler := &api.ShareHandler{DB: pool, Secret: secret} // ✅ now used

	// Router
	r := mux.NewRouter()

	// Public routes
	r.HandleFunc("/register", userHandler.Register).Methods("POST")
	r.HandleFunc("/login", userHandler.Login).Methods("POST")

	// Protected routes
	r.Handle("/files", api.AuthMiddleware(http.HandlerFunc(fileHandler.UploadFile), secret)).Methods("POST")
	r.Handle("/files", api.AuthMiddleware(http.HandlerFunc(fileHandler.GetFiles), secret)).Methods("GET")
	r.Handle("/files/{id}", api.AuthMiddleware(http.HandlerFunc(fileHandler.DownloadFile), secret)).Methods("GET")
	r.Handle("/files/{id}", api.AuthMiddleware(http.HandlerFunc(fileHandler.DeleteFile), secret)).Methods("DELETE")

	r.Handle("/share", api.AuthMiddleware(http.HandlerFunc(fileHandler.ShareFile), secret)).Methods("POST")
	r.Handle("/shared", api.AuthMiddleware(http.HandlerFunc(fileHandler.GetSharedFiles), secret)).Methods("GET")

	r.Handle("/storage", api.AuthMiddleware(http.HandlerFunc(fileHandler.GetStorage), secret)).Methods("GET")

	// Optional: routes using ShareHandler if you extend functionality
	_ = shareHandler // avoids unused error if not yet wired

	// CORS
	headers := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	methods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	origins := handlers.AllowedOrigins([]string{"*"})

	log.Println("🚀 Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", handlers.CORS(headers, methods, origins)(r)))
}
