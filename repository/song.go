package repository

import (
	"github.com/christosgalano/testcontainers-demo/models"
)

// SongRepository defines the methods for managing songs.
type SongRepository interface {
	GetAll() ([]models.Song, error)
	GetByID(id string) (*models.Song, error)
	Create(song *models.Song) error
	Update(song *models.Song) error
	Delete(id string) error
}
