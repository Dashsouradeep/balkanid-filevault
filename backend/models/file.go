package models

import "time"

type File struct {
	ID         int       `json:"id"`
	UserID     int       `json:"user_id"`
	Filename   string    `json:"filename"`
	Filepath   string    `json:"filepath"`
	FileHash   string    `json:"file_hash"`
	RefCount   int       `json:"ref_count"`
	UploadedAt time.Time `json:"uploaded_at"`
}
