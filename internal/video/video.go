package video

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/channyeintun/go-server-for-courses/internal/model"
)

// Scan returns a list of videos found in the given directory.
func Scan(dir string) ([]model.Video, error) {
	var videos []model.Video
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".mp4") {
			relPath, err := filepath.Rel(dir, path)
			if err != nil {
				return err
			}
			relPath = filepath.ToSlash(relPath)
			videos = append(videos, model.Video{
				Name:      info.Name(),
				Path:      relPath,
				Completed: false,
			})
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to scan videos in directory %q: %w", dir, err)
	}
	return videos, nil
}

// GroupBySection takes a list of videos and groups them logically into Sections.
func GroupBySection(videos []model.Video) []model.Section {
	sectionMap := make(map[string][]model.Video)
	var sectionNames []string

	for _, v := range videos {
		parts := strings.Split(v.Path, "/")
		sectionName := "Course Videos"
		if len(parts) >= 2 {
			sectionName = parts[0]
		}
		if _, exists := sectionMap[sectionName]; !exists {
			sectionNames = append(sectionNames, sectionName)
		}
		sectionMap[sectionName] = append(sectionMap[sectionName], v)
	}

	var sections []model.Section
	for _, name := range sectionNames {
		sections = append(sections, model.Section{Name: name, Videos: sectionMap[name]})
	}
	return sections
}

// FindNext finds the current video and the next video in the sequence based on section grouping.
func FindNext(videos []model.Video, currentPath string) (current, next model.Video, found bool) {
	sections := GroupBySection(videos)
	var flat []model.Video
	for _, sec := range sections {
		flat = append(flat, sec.Videos...)
	}

	for i, v := range flat {
		if v.Path == currentPath {
			current = v
			if i+1 < len(flat) {
				return current, flat[i+1], true
			}
			return current, model.Video{}, false
		}
	}
	return model.Video{}, model.Video{}, false
}
