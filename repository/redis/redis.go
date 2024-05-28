package repository

import (
	"context"
	"encoding/json"

	"github.com/go-redis/redis/v8"

	"github.com/christosgalano/testcontainers-demo/model"
)

// RedisSongRepository is a Redis implementation of SongRepository.
type RedisSongRepository struct {
	client *redis.Client
}

// GetAll returns all songs.
func (r *RedisSongRepository) GetAll(ctx context.Context) ([]model.Song, error) {
	keys, err := r.client.Keys(ctx, "*").Result()
	if err != nil {
		return nil, err
	}
	var songs []model.Song
	for _, key := range keys {
		val, err := r.client.Get(ctx, key).Result()
		if err != nil {
			return nil, err
		}
		var song model.Song
		err = json.Unmarshal([]byte(val), &song)
		if err != nil {
			return nil, err
		}
		songs = append(songs, song)
	}
	return songs, nil
}

// GetByID returns a song by ID.
func (r *RedisSongRepository) GetByID(ctx context.Context, id string) (*model.Song, error) {
	val, err := r.client.Get(ctx, id).Result()
	if err != nil {
		return nil, err
	}
	var song model.Song
	err = json.Unmarshal([]byte(val), &song)
	if err != nil {
		return nil, err
	}
	return &song, nil
}

// Create creates a new song.
func (r *RedisSongRepository) Create(ctx context.Context, song *model.Song) (*model.Song, error) {
	songJSON, err := json.Marshal(song)
	if err != nil {
		return nil, err
	}
	err = r.client.Set(ctx, song.ID, songJSON, 0).Err()
	if err != nil {
		return nil, err
	}
	return song, nil
}

// Update updates an existing song.
func (r *RedisSongRepository) Update(ctx context.Context, song *model.Song) (*model.Song, error) {
	return r.Create(ctx, song) // In Redis, update can be done using the same method as create
}

// Delete deletes a song by ID.
func (r *RedisSongRepository) Delete(ctx context.Context, id string) error {
	err := r.client.Del(ctx, id).Err()
	if err != nil {
		return err
	}
	return nil
}
