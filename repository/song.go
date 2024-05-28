package repository

import (
	"github.com/christosgalano/testcontainers-demo/model"
)

// SongRepository defines the methods for managing songs.
type SongRepository interface {
	GetAll() ([]model.Song, error)
	GetByID(id string) (*model.Song, error)
	Create(song *model.Song) error
	Update(song *model.Song) error
	Delete(id string) error
}
