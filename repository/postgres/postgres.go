package repository

import (
	"context"
	"database/sql"

	"github.com/christosgalano/testcontainers-demo/model"
)

// PostgresSongRepository is a PostgreSQL implementation of SongRepository.
type PostgresSongRepository struct {
	db *sql.DB
}

// GetAll returns all songs.
func (r *PostgresSongRepository) GetAll(ctx context.Context) ([]model.Song, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, name, composer FROM songs")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var songs []model.Song
	for rows.Next() {
		var song model.Song
		if err := rows.Scan(&song.ID, &song.Name, &song.Composer); err != nil {
			return nil, err
		}
		songs = append(songs, song)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return songs, nil
}

// GetByID returns a song by ID.
func (r *PostgresSongRepository) GetByID(ctx context.Context, id string) (*model.Song, error) {
	row := r.db.QueryRowContext(ctx, "SELECT id, name, composer FROM songs WHERE id = $1", id)
	var song model.Song
	if err := row.Scan(&song.ID, &song.Name, &song.Composer); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &song, nil
}

// Create creates a new song.
func (r *PostgresSongRepository) Create(ctx context.Context, song *model.Song) (*model.Song, error) {
	row := r.db.QueryRowContext(ctx, "INSERT INTO songs (id, name, composer) VALUES ($1, $2, $3) RETURNING id, name, composer", song.ID, song.Name, song.Composer)
	var newSong model.Song
	err := row.Scan(&newSong.ID, &newSong.Name, &newSong.Composer)
	if err != nil {
		return nil, err
	}
	return &newSong, nil
}

// Update updates an existing song.
func (r *PostgresSongRepository) Update(ctx context.Context, song *model.Song) (*model.Song, error) {
	row := r.db.QueryRowContext(ctx, "UPDATE songs SET name = $1, composer = $2 WHERE id = $3 RETURNING id, name, composer", song.Name, song.Composer, song.ID)
	var updatedSong model.Song
	err := row.Scan(&updatedSong.ID, &updatedSong.Name, &updatedSong.Composer)
	if err != nil {
		return nil, err
	}
	return &updatedSong, nil
}

// Delete deletes a song by ID.
func (r *PostgresSongRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM songs WHERE id = $1", id)
	return err
}
