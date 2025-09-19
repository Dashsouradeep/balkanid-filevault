package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Dashsouradeep/balkanid-filevault/backend/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ShareHandler handles sharing endpoints
type ShareHandler struct {
	DB     *pgxpool.Pool
	Secret string
}

// POST /share → Share a file with another user
func (h *ShareHandler) ShareFile(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromToken(r)
	if !ok {
		http.Error(w, "❌ Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		FileID     int `json:"file_id"`
		TargetUser int `json:"target_user"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "❌ Invalid input", http.StatusBadRequest)
		return
	}

	// Ensure file belongs to sharer
	var ownerID int
	err := h.DB.QueryRow(r.Context(),
		`SELECT user_id FROM files WHERE id=$1`, req.FileID).Scan(&ownerID)
	if err != nil || ownerID != userID {
		http.Error(w, "❌ You don’t own this file", http.StatusForbidden)
		return
	}

	// Insert into shares with conflict handling
	_, err = h.DB.Exec(r.Context(),
		`INSERT INTO shares (file_id, shared_by, target_user, shared_at)
     VALUES ($1, $2, $3, NOW())
     ON CONFLICT (file_id, shared_by, target_user) DO NOTHING`,
		req.FileID, userID, req.TargetUser,
	)

	if err != nil {
		http.Error(w, "DB Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "✅ File shared successfully (or already shared)"})

}

// GET /shared → Files shared with me
func (h *ShareHandler) GetSharedFiles(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromToken(r)
	if !ok {
		http.Error(w, "❌ Unauthorized", http.StatusUnauthorized)
		return
	}

	rows, err := h.DB.Query(context.Background(),
		`SELECT f.id, f.filename, f.filepath, f.file_hash, f.ref_count, f.uploaded_at, s.shared_by
		 FROM shares s
		 JOIN files f ON s.file_id = f.id
		 WHERE s.target_user=$1
		 ORDER BY s.shared_at DESC`, userID)
	if err != nil {
		http.Error(w, "DB Error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var sharedFiles []models.SharedFile
	for rows.Next() {
		var f models.SharedFile
		if err := rows.Scan(
			&f.ID,
			&f.Filename,
			&f.Filepath,
			&f.FileHash,
			&f.RefCount,
			&f.UploadedAt,
			&f.SharedBy,
		); err != nil {
			http.Error(w, "Scan error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		sharedFiles = append(sharedFiles, f)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sharedFiles)

}
