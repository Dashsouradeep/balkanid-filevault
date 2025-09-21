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
	userID, ok := h.getUserID(r)
	if !ok {
		http.Error(w, "❌ Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse multipart form (max 10MB file size here)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "❌ Could not parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "❌ Could not get file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Read file into memory
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "❌ Could not read file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	fileSize := int64(len(fileBytes))

	// Check user quota
	var used, quota int64
	err = h.DB.QueryRow(r.Context(),
		`SELECT used_bytes, quota_bytes FROM user_storage WHERE user_id=$1`, userID,
	).Scan(&used, &quota)

	if err == sql.ErrNoRows {
		// initialize record if missing (100 MB quota default)
		_, err = h.DB.Exec(r.Context(),
			`INSERT INTO user_storage (user_id, used_bytes, quota_bytes) VALUES ($1, 0, 104857600)`,
			userID)
		if err != nil {
			http.Error(w, "DB Error (init storage): "+err.Error(), http.StatusInternalServerError)
			return
		}
		used, quota = 0, 104857600
	} else if err != nil {
		http.Error(w, "DB Error (check quota): "+err.Error(), http.StatusInternalServerError)
		return
	}

	if used+fileSize > quota {
		http.Error(w, "❌ Storage quota exceeded", http.StatusForbidden)
		return
	}

	// Compute SHA-256 hash
	hash := sha256.Sum256(fileBytes)
	fileHash := hex.EncodeToString(hash[:])

	// Save file physically (if not already present)
	uploadsDir := "uploads"
	filePath := filepath.Join(uploadsDir, handler.Filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		dst, err := CreateFile(filePath)
		if err != nil {
			http.Error(w, "❌ Could not save file: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		if _, err := dst.Write(fileBytes); err != nil {
			http.Error(w, "❌ Could not write file: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	var fileID int
	rows, err := h.DB.Query(r.Context(),
		`WITH upsert AS (
		 INSERT INTO files (user_id, filename, filepath, file_hash, ref_count, uploaded_at)
		 VALUES ($1, $2, $3, $4, 1, NOW())
		 ON CONFLICT (file_hash) DO UPDATE 
		 SET ref_count = files.ref_count + 1
		 RETURNING id
	 )
	 SELECT id FROM upsert`,
		userID, handler.Filename, filePath, fileHash,
	)
	if err != nil {
		http.Error(w, "DB Error (insert file query): "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&fileID); err != nil {
			http.Error(w, "DB Error (scan id): "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(w, "DB Error: no id returned from insert", http.StatusInternalServerError)
		return
	}

	// Update user storage usage
	_, err = h.DB.Exec(r.Context(),
		`UPDATE user_storage SET used_bytes = used_bytes + $1 WHERE user_id=$2`,
		fileSize, userID,
	)
	if err != nil {
		http.Error(w, "DB Error (update storage): "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Response
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprintf(
		`{"message":"✅ File uploaded (deduplication + quota enforced)","file_id":%d}`, fileID,
	)))
}

// GetFiles - list uploaded files for the logged-in user
func (h *FileHandler) GetFiles(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.getUserID(r)
	if !ok {
		http.Error(w, "❌ Unauthorized", http.StatusUnauthorized)
		return
	}

	rows, err := h.DB.Query(r.Context(),
		`SELECT id,
		        COALESCE(user_id, 0) AS user_id,
		        filename,
		        COALESCE(filepath, '') AS filepath,
		        file_hash,
		        COALESCE(ref_count, 0) AS ref_count,
		        uploaded_at
		 FROM files
		 WHERE user_id = $1
		 ORDER BY uploaded_at DESC`, userID)
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
			&f.FileHash,
			&f.RefCount,
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

// DownloadFile - download a file by ID (owner or shared)
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

	// Check ownership or shared access
	if ownerID != userID {
		var count int
		err := h.DB.QueryRow(r.Context(),
			`SELECT COUNT(*) FROM shares WHERE file_id=$1 AND target_user=$2`,
			fileID, userID).Scan(&count)
		if err != nil || count == 0 {
			http.Error(w, "❌ Forbidden", http.StatusForbidden)
			return
		}
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	http.ServeFile(w, r, filepath.Clean(filePath))
}

// ShareFile - share a file with another user
func (h *FileHandler) ShareFile(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.getUserID(r)
	if !ok {
		http.Error(w, "❌ Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		FileID     int    `json:"file_id"`
		TargetUser int    `json:"target_user"`
		ShareType  string `json:"share_type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Default share_type if empty
	if req.ShareType == "" {
		req.ShareType = "read"
	}

	// Insert into shares table
	_, err := h.DB.Exec(r.Context(),
		`INSERT INTO shares (file_id, owner_id, target_user, share_type, shared_at)
         VALUES ($1, $2, $3, $4, NOW())`,
		req.FileID, userID, req.TargetUser, req.ShareType,
	)
	if err != nil {
		http.Error(w, "DB Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message":"✅ File shared successfully"}`))
}

// GetSharedFiles - list files shared *with* the logged-in user
func (h *FileHandler) GetSharedFiles(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.getUserID(r)
	if !ok {
		http.Error(w, "❌ Unauthorized", http.StatusUnauthorized)
		return
	}

	rows, err := h.DB.Query(r.Context(),
		`SELECT f.id, f.user_id, f.filename, f.filepath, f.file_hash, f.ref_count, f.uploaded_at, s.share_type, s.owner_id
         FROM files f
         JOIN shares s ON f.id = s.file_id
         WHERE s.target_user=$1
         ORDER BY s.shared_at DESC`, userID)
	if err != nil {
		http.Error(w, "DB Error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type SharedFile struct {
		models.File
		OwnerID   int    `json:"owner_id"`
		ShareType string `json:"share_type"`
	}

	var files []SharedFile
	for rows.Next() {
		var sf SharedFile
		if err := rows.Scan(
			&sf.ID, &sf.UserID, &sf.Filename, &sf.Filepath,
			&sf.FileHash, &sf.RefCount, &sf.UploadedAt,
			&sf.ShareType, &sf.OwnerID,
		); err != nil {
			http.Error(w, "Scan Error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		files = append(files, sf)
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

// GET /storage → check quota usage
func (h *FileHandler) GetStorage(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromToken(r)
	if !ok {
		http.Error(w, "❌ Unauthorized", http.StatusUnauthorized)
		return
	}

	var used, quota int64
	err := h.DB.QueryRow(r.Context(),
		`SELECT used_bytes, quota_bytes FROM user_storage WHERE user_id=$1`, userID,
	).Scan(&used, &quota)

	if err == sql.ErrNoRows {
		http.Error(w, "❌ No storage record found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "DB Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"used_bytes":   used,
		"quota_bytes":  quota,
		"percent_used": float64(used) / float64(quota) * 100,
	})
}

// DeleteFile - delete a file owned by the user
func (h *FileHandler) DeleteFile(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.getUserID(r)
	if !ok {
		http.Error(w, "❌ Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	fileID := vars["id"]

	var ownerID int
	var filePath string
	var refCount int
	err := h.DB.QueryRow(r.Context(),
		`SELECT user_id, filepath, ref_count FROM files WHERE id=$1`, fileID).
		Scan(&ownerID, &filePath, &refCount)
	if err == sql.ErrNoRows {
		http.Error(w, "❌ File not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "DB Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Ensure user owns file
	if ownerID != userID {
		http.Error(w, "❌ Forbidden", http.StatusForbidden)
		return
	}

	if refCount > 1 {
		// Just decrement ref_count
		_, err = h.DB.Exec(r.Context(),
			`UPDATE files SET ref_count = ref_count - 1 WHERE id=$1`, fileID)
		if err != nil {
			http.Error(w, "DB Error: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		// Remove DB row and physical file
		_, err = h.DB.Exec(r.Context(),
			`DELETE FROM files WHERE id=$1`, fileID)
		if err != nil {
			http.Error(w, "DB Error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			http.Error(w, "❌ Could not delete file: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "✅ File deleted successfully"})
}
