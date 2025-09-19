package api

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Dashsouradeep/balkanid-filevault/backend/models"
	"github.com/Dashsouradeep/balkanid-filevault/backend/utils"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
)

// FileHandler handles file-related routes
type FileHandler struct {
	DB     *pgxpool.Pool
	Secret string
}

// helper: extract user ID from JWT token
func (h *FileHandler) getUserID(r *http.Request) (int, bool) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return 0, false
	}

	tokenStr := utils.ExtractToken(authHeader)
	claims, err := utils.ValidateAndGetClaims(tokenStr, h.Secret)
	if err != nil {
		return 0, false
	}

	userID, ok := claims["user_id"].(float64)
	if !ok {
		return 0, false
	}

	return int(userID), true
}

// UploadFile - handle file uploads with deduplication
func (h *FileHandler) UploadFile(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromToken(r)
	if !ok {
		http.Error(w, "❌ Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB limit
		http.Error(w, "❌ Could not parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "❌ Could not get file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Compute SHA-256 hash
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		http.Error(w, "❌ Could not hash file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	fileHash := hex.EncodeToString(hasher.Sum(nil))

	// Reset pointer
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		http.Error(w, "❌ Could not reset file pointer: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Save file physically (only if not already present)
	filePath := "uploads/" + handler.Filename
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		dst, err := CreateFile(filePath)
		if err != nil {
			http.Error(w, "❌ Could not save file: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			http.Error(w, "❌ Could not write file: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Insert or update ref_count in DB
	var fileID int
	err = h.DB.QueryRow(r.Context(),
		`INSERT INTO files (user_id, filename, filepath, file_hash, ref_count, uploaded_at)
		 VALUES ($1, $2, $3, $4, 1, NOW())
		 ON CONFLICT (file_hash) DO UPDATE 
		 SET ref_count = files.ref_count + 1
		 RETURNING id`,
		userID, handler.Filename, filePath, fileHash,
	).Scan(&fileID)

	if err != nil {
		http.Error(w, "DB Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Response
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprintf(`{"message":"✅ File uploaded (deduplication applied)","file_id":%d}`, fileID)))
}

// GetFiles - list uploaded files for the logged-in user (with hash + ref_count)
// GetFiles - list uploaded files for the logged-in user
func (h *FileHandler) GetFiles(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromToken(r)
	if !ok {
		http.Error(w, "❌ Unauthorized", http.StatusUnauthorized)
		return
	}

	rows, err := h.DB.Query(r.Context(),
		`SELECT id, user_id, filename, filepath, file_hash, ref_count, uploaded_at
         FROM files WHERE user_id=$1 ORDER BY uploaded_at DESC`, userID)
	if err != nil {
		http.Error(w, "DB Error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var files []models.File
	for rows.Next() {
		var f models.File
		if err := rows.Scan(
			&f.ID,
			&f.UserID,
			&f.Filename,
			&f.Filepath,
			&f.FileHash, // ✅ new field
			&f.RefCount, // ✅ new field
			&f.UploadedAt,
		); err != nil {
			http.Error(w, "Scan Error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		files = append(files, f)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

// DownloadFile - download a file by ID
func (h *FileHandler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.getUserID(r)
	if !ok {
		http.Error(w, "❌ Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	fileID := vars["id"]

	var ownerID int
	var filePath, fileName string
	err := h.DB.QueryRow(r.Context(),
		`SELECT user_id, filename, filepath FROM files WHERE id=$1`, fileID).
		Scan(&ownerID, &fileName, &filePath)

	if err == sql.ErrNoRows {
		http.Error(w, "❌ File not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "DB Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if ownerID != userID {
		http.Error(w, "❌ Forbidden", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	http.ServeFile(w, r, filepath.Clean(filePath))
}

// CreateFile ensures the uploads directory exists and creates the file
func CreateFile(path string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return nil, err
	}
	return os.Create(path)
}
