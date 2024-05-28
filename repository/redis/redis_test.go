package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/testcontainers/testcontainers-go"
	cr "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
	"gotest.tools/v3/assert"

	"github.com/christosgalano/testcontainers-demo/model"
)

func setupTestRedisRepository(ctx context.Context) (*RedisSongRepository, func(), error) {
	// Start a Redis container
	container, err := cr.RunContainer(
		ctx,
		testcontainers.WithImage("redis:7"),
		cr.WithLogLevel(cr.LogLevelVerbose),
		testcontainers.WithWaitStrategy(wait.ForListeningPort("6379/tcp")),
	)
	if err != nil {
		return nil, nil, err
	}
	endpoint, err := container.Endpoint(ctx, "")
	if err != nil {
		return nil, nil, err
	}
	log.Printf("redis container endpoint: %s", endpoint)

	// Create a Redis client
	client := redis.NewClient(
		&redis.Options{
			Addr: endpoint,
		},
	)
	_, err = client.Ping(ctx).Result()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to ping Redis: %w", err)
	}
	log.Printf("created redis client")

	// Create and a RedisSongRepository
	repo := &RedisSongRepository{client: client}
	log.Printf("created redis song repository")

	// Initialize the Redis store with test data
	initialSongs, err := os.ReadFile("./testdata/songs.json")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read songs.json: %w", err)
	}
	var songs []model.Song
	if err := json.Unmarshal(initialSongs, &songs); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal songs: %w", err)
	}
	for _, s := range songs {
		song, err := json.Marshal(s)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal song: %w", err)
		}
		if err := client.Set(ctx, s.ID, song, 0).Err(); err != nil {
			return nil, nil, fmt.Errorf("failed to set song: %w", err)
		}
	}
	log.Printf("initialized redis store")

	// Return the repository and a cleanup function
	cleanup := func() {
		if err := container.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}

	return repo, cleanup, nil
}

func TestRedisSongRepository_GetAll(t *testing.T) {
	ctx := context.Background()

	repo, cleanup, err := setupTestRedisRepository(ctx)
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

	expectedSongs := make(map[string]model.Song)
	for i := 1; i <= len(songs); i++ {
		expectedSongs[fmt.Sprintf("%d", i)] = model.Song{
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

func TestRedisSongRepository_GetByID(t *testing.T) {
	ctx := context.Background()

	repo, cleanup, err := setupTestRedisRepository(ctx)
	if err != nil {
		t.Fatalf("failed to setup test: %s", err)
	}
	defer cleanup()

	song, err := repo.GetByID(ctx, "1")
	if err != nil {
		t.Fatalf("failed to get song by ID: %s", err)
	}

	expectedSong := model.Song{
		ID:       "1",
		Name:     "Song 1",
		Composer: "Composer 1",
	}
	assert.Equal(t, *song, expectedSong)

	nonExistentSong, err := repo.GetByID(ctx, "4")
	if err == nil {
		t.Fatalf("failed to return nil for non-existent song: %s", err)
	}
	assert.Equal(t, nonExistentSong, (*model.Song)(nil))
}

func TestRedisSongRepository_Create(t *testing.T) {
	ctx := context.Background()

	repo, cleanup, err := setupTestRedisRepository(ctx)
	if err != nil {
		t.Fatalf("failed to setup test: %s", err)
	}
	defer cleanup()

	song := &model.Song{
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

func TestRedisSongRepository_Update(t *testing.T) {
	ctx := context.Background()

	repo, cleanup, err := setupTestRedisRepository(ctx)
	if err != nil {
		t.Fatalf("failed to setup test: %s", err)
	}
	defer cleanup()

	song := &model.Song{
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

func TestRedisSongRepository_Delete(t *testing.T) {
	ctx := context.Background()

	repo, cleanup, err := setupTestRedisRepository(ctx)
	if err != nil {
		t.Fatalf("failed to setup test: %s", err)
	}
	defer cleanup()

	err = repo.Delete(ctx, "1")
	if err != nil {
		t.Fatalf("failed to delete song: %s", err)
	}

	song, err := repo.GetByID(ctx, "1")
	if err == nil {
		t.Fatalf("failed to return nil for deleted song: %s", err)
	}
	assert.Equal(t, song, (*model.Song)(nil))

	err = repo.Delete(ctx, "4")
	if err != nil {
		t.Fatalf("failed to delete non-existent song: %s", err)
	}
}
