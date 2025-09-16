package models

import "time"

type File struct {
	ID         int       `json:"id"`
	UserID     int       `json:"user_id"`
	FileHash   string    `json:"file_hash"`
	Filename   string    `json:"filename"`
	MimeType   string    `json:"mime_type"`
	UploadedAt time.Time `json:"uploaded_at"`
	IsDeleted  bool      `json:"is_deleted"`
}
