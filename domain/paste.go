package domain

import (
	"context"
	"time"
)

// Paste entity
type Paste struct {
	Title        string
	ShortURLPath string
	TextData     string
	StorageURL   string
	AuthorUserID string
	Private      bool
	ExpiredAt    time.Time
}

// PasteService interface is an abstraction of a service
// that provides resource management of paste resources
// and controls the flow of paste Read() and Write()
type PasteService interface {
	WritePaste(ctx context.Context, paste Paste) (string, error)
	ReadPaste(ctx context.Context, shortURLPath string) (Paste, error)
}

// PasteRepository interface represent mongoDB database
// and handles Paste resources
type PasteRepository interface {
	GetPaste(ctx context.Context, shortURLPath string) (Paste, error)
	DeletePaste(ctx context.Context, shortURLPath string) error
	CreatePaste(ctx context.Context, paste Paste) error
}
