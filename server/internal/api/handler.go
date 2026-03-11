package api

import (
	"wayback/internal/database"
	"wayback/internal/storage"
)

type Handler struct {
	dedup   *storage.Deduplicator
	db      *database.DB
	dataDir string
}

func NewHandler(dedup *storage.Deduplicator, db *database.DB, dataDir string) *Handler {
	return &Handler{
		dedup:   dedup,
		db:      db,
		dataDir: dataDir,
	}
}
