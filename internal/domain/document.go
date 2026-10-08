package domain

import (
	"time"

	"github.com/google/uuid"
)

type Document struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	SourceURI   string    `json:"source_uri"`
	MimeType    string    `json:"mime_type"`
	TotalChunks int       `json:"total_chunks"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func NewDocument(title, sourceURI, mimeType string) *Document {
	now := time.Now().UTC()
	return &Document{
		ID:          uuid.New(),
		Title:       title,
		SourceURI:   sourceURI,
		MimeType:    mimeType,
		TotalChunks: 0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}
