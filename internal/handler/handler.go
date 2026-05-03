package handler

import (
	"fmt"
	"html/template"
	"net/http"
	"sync"

	"github.com/channyeintun/go-server-for-courses/internal/model"
	"github.com/channyeintun/go-server-for-courses/internal/progress"
	"github.com/channyeintun/go-server-for-courses/internal/video"
)

// AppHandler contains the application state required by HTTP handlers.
type AppHandler struct {
	mu            sync.Mutex
	videos        []model.Video
	progressStore *progress.Store
	autoPlay      bool
	tmpl          *template.Template
}

// NewAppHandler initializes the handler with the initial videos and state.
func NewAppHandler(videos []model.Video, progressStore *progress.Store, tmpl *template.Template) *AppHandler {
	return &AppHandler{
		videos:        videos,
		progressStore: progressStore,
		autoPlay:      false,
		tmpl:          tmpl,
	}
}

// Home renders the main landing page.
func (h *AppHandler) Home(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	sections := video.GroupBySection(h.videos)
	autoPlay := h.autoPlay
	h.mu.Unlock()

	data := struct {
		Sections     []model.Section
		CurrentVideo *model.Video
		AutoPlay     bool
	}{
		Sections: sections,
		AutoPlay: autoPlay,
	}
	if err := h.tmpl.ExecuteTemplate(w, "index", data); err != nil {
		http.Error(w, fmt.Sprintf("template rendering failed: %v", err), http.StatusInternalServerError)
	}
}

// Toggle flips the completed state of a video.
func (h *AppHandler) Toggle(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")

	h.mu.Lock()
	defer h.mu.Unlock()

	var toggledVideo model.Video
	for i, v := range h.videos {
		if v.Path == path {
			h.videos[i].Completed = !h.videos[i].Completed
			toggledVideo = h.videos[i]
			break
		}
	}
	if err := h.progressStore.Save(h.videos); err != nil {
		http.Error(w, fmt.Sprintf("failed to save progress: %v", err), http.StatusInternalServerError)
		return
	}

	if err := h.tmpl.ExecuteTemplate(w, "toggle-btn", map[string]interface{}{
		"Video": toggledVideo,
	}); err != nil {
		http.Error(w, fmt.Sprintf("template rendering failed: %v", err), http.StatusInternalServerError)
	}
}

// Play returns the player HTML for the requested video.
func (h *AppHandler) Play(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")

	h.mu.Lock()
	var current model.Video
	for _, v := range h.videos {
		if v.Path == path {
			current = v
			break
		}
	}
	sections := video.GroupBySection(h.videos)
	autoPlay := h.autoPlay
	h.mu.Unlock()

	data := struct {
		CurrentVideo *model.Video
		AutoPlay     bool
	}{
		CurrentVideo: &current,
		AutoPlay:     autoPlay,
	}

	if err := h.tmpl.ExecuteTemplate(w, "player", data); err != nil {
		http.Error(w, fmt.Sprintf("template rendering failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("\n<div id=\"playlist-container\" hx-swap-oob=\"true\" class=\"w-full md:w-[400px] bg-gray-900 border-l border-gray-800 flex flex-col h-[50vh] md:h-screen shrink-0\">\n"))
	if err := h.tmpl.ExecuteTemplate(w, "playlist", struct {
		Sections     []model.Section
		CurrentVideo *model.Video
	}{
		Sections:     sections,
		CurrentVideo: &current,
	}); err != nil {
		http.Error(w, fmt.Sprintf("template rendering failed: %v", err), http.StatusInternalServerError)
	}
	w.Write([]byte("\n</div>"))
}

// Autoplay toggles the autoplay feature.
func (h *AppHandler) Autoplay(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	h.autoPlay = !h.autoPlay
	autoPlay := h.autoPlay
	h.mu.Unlock()

	if err := h.tmpl.ExecuteTemplate(w, "autoplay-btn", struct{ AutoPlay bool }{AutoPlay: autoPlay}); err != nil {
		http.Error(w, fmt.Sprintf("template rendering failed: %v", err), http.StatusInternalServerError)
	}
}

// Ended handles the logic when a video finishes playing.
func (h *AppHandler) Ended(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")

	h.mu.Lock()
	defer h.mu.Unlock()

	var current model.Video
	var next model.Video
	found := false

	var flatVideos []model.Video
	sections := video.GroupBySection(h.videos)
	for _, sec := range sections {
		flatVideos = append(flatVideos, sec.Videos...)
	}

	for i, v := range flatVideos {
		if v.Path == path {
			current = v
			if i+1 < len(flatVideos) {
				next = flatVideos[i+1]
				found = true
			}
			break
		}
	}

	for i, v := range h.videos {
		if v.Path == path {
			h.videos[i].Completed = true
			break
		}
	}

	if err := h.progressStore.Save(h.videos); err != nil {
		http.Error(w, fmt.Sprintf("failed to save progress: %v", err), http.StatusInternalServerError)
		return
	}

	var targetVideo *model.Video
	if h.autoPlay && found {
		targetVideo = &next
	} else {
		current.Completed = true
		targetVideo = &current
	}

	data := struct {
		CurrentVideo *model.Video
		AutoPlay     bool
	}{
		CurrentVideo: targetVideo,
		AutoPlay:     h.autoPlay,
	}

	if err := h.tmpl.ExecuteTemplate(w, "player", data); err != nil {
		http.Error(w, fmt.Sprintf("template rendering failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("\n<div id=\"playlist-container\" hx-swap-oob=\"true\" class=\"w-full md:w-[400px] bg-gray-900 border-l border-gray-800 flex flex-col h-[50vh] md:h-screen shrink-0\">\n"))
	if err := h.tmpl.ExecuteTemplate(w, "playlist", struct {
		Sections     []model.Section
		CurrentVideo *model.Video
	}{
		Sections:     video.GroupBySection(h.videos),
		CurrentVideo: targetVideo,
	}); err != nil {
		http.Error(w, fmt.Sprintf("template rendering failed: %v", err), http.StatusInternalServerError)
	}
	w.Write([]byte("\n</div>"))
}
