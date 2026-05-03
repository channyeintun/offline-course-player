package progress

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/channyeintun/go-server-for-courses/internal/model"
)

// Store handles saving and loading video progress.
type Store struct {
	filePath string
}

// NewStore initializes a new Store with the target file path.
func NewStore(filePath string) *Store {
	return &Store{filePath: filePath}
}

// Load reads the progress file and updates the given videos' completion status.
func (s *Store) Load(videos []model.Video) ([]model.Video, error) {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return videos, nil // No progress saved yet
		}
		return nil, fmt.Errorf("failed to read progress file: %w", err)
	}

	var savedVideos []model.Video
	if err := json.Unmarshal(data, &savedVideos); err != nil {
		return nil, fmt.Errorf("failed to parse progress data: %w", err)
	}

	for i := range videos {
		for _, saved := range savedVideos {
			if videos[i].Path == saved.Path {
				videos[i].Completed = saved.Completed
				break
			}
		}
	}
	return videos, nil
}

// Save writes the current videos' progress to the file.
func (s *Store) Save(videos []model.Video) error {
	data, err := json.Marshal(videos)
	if err != nil {
		return fmt.Errorf("failed to marshal progress: %w", err)
	}

	if err := os.WriteFile(s.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write progress file: %w", err)
	}
	return nil
}
