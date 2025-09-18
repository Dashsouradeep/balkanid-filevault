package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Dashsouradeep/balkanid-filevault/backend/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

// FileHandler struct
type FileHandler struct {
	DB *pgxpool.Pool
}

// UploadFile - handle file uploads
func (h *FileHandler) UploadFile(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromToken(r)
	if !ok {
		http.Error(w, "❌ Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		http.Error(w, "❌ Could not parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "❌ Could not get file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Save file to uploads/ directory
	filePath := "uploads/" + handler.Filename
	dst, err := CreateFile(filePath)
	if err != nil {
		http.Error(w, "❌ Could not save file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	_, err = dst.ReadFrom(file)
	if err != nil {
		http.Error(w, "❌ Could not write file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Insert metadata into DB
	_, err = h.DB.Exec(r.Context(),
		`INSERT INTO files (user_id, filename, filepath, uploaded_at)
		 VALUES ($1, $2, $3, NOW())`,
		userID, handler.Filename, filePath,
	)
	if err != nil {
		http.Error(w, "DB Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message":"✅ File uploaded successfully","user_id":` + fmt.Sprint(userID) + `}`))
}

// GetFiles - list uploaded files for the logged-in user
func (h *FileHandler) GetFiles(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromToken(r)
	if !ok {
		http.Error(w, "❌ Unauthorized", http.StatusUnauthorized)
		return
	}

	rows, err := h.DB.Query(r.Context(),
		`SELECT id, user_id, filename, filepath, uploaded_at 
		 FROM files WHERE user_id=$1 ORDER BY uploaded_at DESC`, userID)
	if err != nil {
		http.Error(w, "DB Error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var files []models.File
	for rows.Next() {
		var f models.File
		if err := rows.Scan(&f.ID, &f.UserID, &f.Filename, &f.Filepath, &f.UploadedAt); err != nil {
			http.Error(w, "Scan Error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		files = append(files, f)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

// CreateFile ensures the uploads directory exists and creates the file
func CreateFile(path string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return nil, err
	}
	return os.Create(path)
}
