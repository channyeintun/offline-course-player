package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/channyeintun/go-server-for-courses/internal/handler"
	"github.com/channyeintun/go-server-for-courses/internal/progress"
	"github.com/channyeintun/go-server-for-courses/internal/video"
	"github.com/channyeintun/go-server-for-courses/internal/view"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func run() error {
	videosDir := "videos"
	dataFile := "progress.json"

	if _, err := os.Stat(videosDir); os.IsNotExist(err) {
		return fmt.Errorf("videos directory '%s' does not exist", videosDir)
	}

	videos, err := video.Scan(videosDir)
	if err != nil {
		return fmt.Errorf("failed to load videos: %w", err)
	}

	store := progress.NewStore(dataFile)
	videos, err = store.Load(videos)
	if err != nil {
		return fmt.Errorf("failed to load progress: %w", err)
	}

	tmpl, err := view.ParseTemplates()
	if err != nil {
		return fmt.Errorf("failed to parse templates: %w", err)
	}

	appHandler := handler.NewAppHandler(videos, store, tmpl)

	mux := http.NewServeMux()
	mux.HandleFunc("/", appHandler.Home)
	mux.HandleFunc("/toggle", appHandler.Toggle)
	mux.HandleFunc("/play", appHandler.Play)
	mux.HandleFunc("/autoplay", appHandler.Autoplay)
	mux.HandleFunc("/ended", appHandler.Ended)
	mux.Handle("/videos/", http.StripPrefix("/videos/", http.FileServer(http.Dir(videosDir))))

	port := ":8080"
	fmt.Printf("Server is running on http://localhost%s\n", port)
	
	if err := http.ListenAndServe(port, mux); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	
	return nil
}
