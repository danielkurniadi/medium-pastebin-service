package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/iqdf/pastebin-service/domain"
)

const timeout = time.Second * 10
const expired = 300 * time.Second

// PasteService serves as an API that provides resource management
// of paste resources and controls the flow of paste Read() and Write()
type PasteService struct {
	pasteRepo domain.PasteRepository
}

// NewPasteService creates new paste service
func NewPasteService(pasteRepo domain.PasteRepository) *PasteService {
	return &PasteService{
		pasteRepo: pasteRepo,
	}
}

// WritePaste creates new paste with writen content/data
func (service *PasteService) WritePaste(ctx context.Context, paste domain.Paste) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	shortURLPath, _ := generateRandomURL(8)
	paste.ShortURLPath = shortURLPath
	paste.ExpiredAt = time.Now().Add(expired)

	return shortURLPath, service.pasteRepo.CreatePaste(ctx, paste)
}

func generateRandomURL(n int) (string, error) {
	b := make([]byte, n/2)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// ReadPaste retrieves previously created paste
// that hasn't expired yet.
func (service *PasteService) ReadPaste(ctx context.Context, shortURLPath string) (domain.Paste, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	paste, err := service.pasteRepo.GetPaste(ctx, shortURLPath)
	return paste, err
}
