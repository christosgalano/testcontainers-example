package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gotest.tools/assert"

	"github.com/christosgalano/testcontainers-demo/models"
)

func setupTestPostgresRepository(ctx context.Context) (*PostgresSongRepository, func(), error) {
	username, password, database := "user", "password", "songs"

	// Start a PostgreSQL container
	container, err := postgres.RunContainer(
		ctx,
		testcontainers.WithImage("postgres:16"),
		postgres.WithDatabase(database),
		postgres.WithUsername(username),
		postgres.WithPassword(password),
		postgres.WithInitScripts(filepath.Join("testdata", "init-song-db.sql")),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
		),
	)
	if err != nil {
		return nil, nil, err
	}
	endpoint, err := container.Endpoint(ctx, "")
	if err != nil {
		return nil, nil, err
	}
	log.Printf("postgres container endpoint: %s", endpoint)

	// Create a PostgreSQL client
	db, err := sql.Open("postgres", fmt.Sprintf(
		"postgresql://%s:%s@%s/%s?sslmode=disable",
		username, password, endpoint, database,
	))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open database: %w", err)
	}
	log.Printf("created postgres client")

	// Create and a PostgresSongRepository
	repo := &PostgresSongRepository{db: db}
	log.Printf("created postgres song repository")

	// Return the repository and a cleanup function
	cleanup := func() {
		if err := container.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}

	return repo, cleanup, nil
}

func TestPostgresSongRepository_GetAll(t *testing.T) {
	ctx := context.Background()

	repo, cleanup, err := setupTestPostgresRepository(ctx)
	if err != nil {
		t.Fatalf("failed to setup test: %s", err)
	}
	defer cleanup()

	songs, err := repo.GetAll(ctx)
	if err != nil {
		t.Fatalf("failed to get all songs: %s", err)
	}

	if len(songs) != 3 {
		t.Fatalf("expected 3 songs, got %d", len(songs))
	}

	expectedSongs := make(map[string]models.Song)
	for i := 1; i <= len(songs); i++ {
		expectedSongs[fmt.Sprintf("%d", i)] = models.Song{
			ID:       fmt.Sprintf("%d", i),
			Name:     fmt.Sprintf("Song %d", i),
			Composer: fmt.Sprintf("Composer %d", i),
		}
	}

	for _, s := range songs {
		expectedSong, ok := expectedSongs[s.ID]
		if !ok {
			t.Errorf("Unexpected song: %+v", s)
			continue
		}
		if s.Name != expectedSong.Name || s.Composer != expectedSong.Composer {
			t.Errorf("Expected song %+v, got %+v", expectedSong, s)
		}
	}
}

func TestPostgresSongRepository_GetByID(t *testing.T) {
	ctx := context.Background()

	repo, cleanup, err := setupTestPostgresRepository(ctx)
	if err != nil {
		t.Fatalf("failed to setup test: %s", err)
	}
	defer cleanup()

	song, err := repo.GetByID(ctx, "1")
	if err != nil {
		t.Fatalf("failed to get song by ID: %s", err)
	}

	expectedSong := models.Song{
		ID:       "1",
		Name:     "Song 1",
		Composer: "Composer 1",
	}
	assert.Equal(t, *song, expectedSong)

	nonExistentSong, err := repo.GetByID(ctx, "4")
	if err != nil {
		t.Fatalf("failed to return nil for non-existent song: %s", err)
	}
	assert.Equal(t, nonExistentSong, (*models.Song)(nil))
}

func TestPostgresSongRepository_Create(t *testing.T) {
	ctx := context.Background()

	repo, cleanup, err := setupTestPostgresRepository(ctx)
	if err != nil {
		t.Fatalf("failed to setup test: %s", err)
	}
	defer cleanup()

	song := &models.Song{
		ID:       "4",
		Name:     "Song 4",
		Composer: "Composer 4",
	}
	createdSong, err := repo.Create(ctx, song)
	if err != nil {
		t.Fatalf("failed to create song: %s", err)
	}

	assert.Equal(t, *createdSong, *song)
}

func TestPostgresSongRepository_Update(t *testing.T) {
	ctx := context.Background()

	repo, cleanup, err := setupTestPostgresRepository(ctx)
	if err != nil {
		t.Fatalf("failed to setup test: %s", err)
	}
	defer cleanup()

	song := &models.Song{
		ID:       "1",
		Name:     "Updated Song 1",
		Composer: "Updated Composer 1",
	}
	updatedSong, err := repo.Update(ctx, song)
	if err != nil {
		t.Fatalf("failed to update song: %s", err)
	}

	assert.Equal(t, *updatedSong, *song)
}

func TestPostgresSongRepository_Delete(t *testing.T) {
	ctx := context.Background()

	repo, cleanup, err := setupTestPostgresRepository(ctx)
	if err != nil {
		t.Fatalf("failed to setup test: %s", err)
	}
	defer cleanup()

	err = repo.Delete(ctx, "1")
	if err != nil {
		t.Fatalf("failed to delete song: %s", err)
	}

	song, err := repo.GetByID(ctx, "1")
	if err != nil {
		t.Fatalf("failed to get song by ID: %s", err)
	}
	assert.Equal(t, song, (*models.Song)(nil))
}
