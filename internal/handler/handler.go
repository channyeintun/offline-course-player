package handler

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
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
	playbackRate  float64
	tmpl          *template.Template
}

// NewAppHandler initializes the handler with the initial videos and state.
func NewAppHandler(videos []model.Video, progressStore *progress.Store, tmpl *template.Template) *AppHandler {
	return &AppHandler{
		videos:        videos,
		progressStore: progressStore,
		autoPlay:      false,
		playbackRate:  1.0,
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
		Sections       []model.Section
		CurrentVideo   *model.Video
		AutoPlay       bool
		Speed          float64
		ShouldAutoplay bool
	}{
		Sections:       sections,
		AutoPlay:       autoPlay,
		Speed:          h.playbackRate,
		ShouldAutoplay: false,
	}
	if err := h.tmpl.ExecuteTemplate(w, "index", data); err != nil {
		http.Error(w, fmt.Sprintf("template rendering failed: %v", err), http.StatusInternalServerError)
	}
}

// Toggle flips the completed state of a video.
func (h *AppHandler) Toggle(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")

	v, err := h.toggleVideoCompletion(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := h.tmpl.ExecuteTemplate(w, "toggle-btn", map[string]interface{}{
		"Video": v,
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
	autoPlay := h.autoPlay
	h.mu.Unlock()

	data := struct {
		CurrentVideo   *model.Video
		AutoPlay       bool
		Speed          float64
		ShouldAutoplay bool
	}{
		CurrentVideo:   &current,
		AutoPlay:       autoPlay,
		Speed:          h.playbackRate,
		ShouldAutoplay: true,
	}

	if err := h.tmpl.ExecuteTemplate(w, "player", data); err != nil {
		http.Error(w, fmt.Sprintf("template rendering failed: %v", err), http.StatusInternalServerError)
		return
	}
}

// Autoplay toggles the autoplay feature.
func (h *AppHandler) Autoplay(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	h.autoPlay = !h.autoPlay
	autoPlay := h.autoPlay
	h.mu.Unlock()

	if err := h.tmpl.ExecuteTemplate(w, "autoplay-btn", struct {
		AutoPlay bool
		Speed    float64
	}{
		AutoPlay: autoPlay,
		Speed:    h.playbackRate,
	}); err != nil {
		http.Error(w, fmt.Sprintf("template rendering failed: %v", err), http.StatusInternalServerError)
	}
}

// Ended handles the logic when a video finishes playing.
func (h *AppHandler) Ended(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	speed := parseSpeed(r.FormValue("speed"))

	state, err := h.updateProgress(path, true, speed)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Determine what to show next
	targetVideo := state.Current
	shouldAutoplay := state.AutoPlay && state.Found
	if shouldAutoplay {
		targetVideo = state.Next
	}

	h.renderPlayerResponse(w, targetVideo, state.AutoPlay, state.PlaybackRate, shouldAutoplay)
	h.renderToggleOOB(w, state.Current)
}




type ProgressState struct {
	Current      model.Video
	Next         model.Video
	Found        bool
	AutoPlay     bool
	PlaybackRate float64
}

func (h *AppHandler) updateProgress(path string, completed bool, speed float64) (ProgressState, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.playbackRate = speed

	// Mark as completed
	for i, v := range h.videos {
		if v.Path == path {
			h.videos[i].Completed = completed
			break
		}
	}

	if err := h.progressStore.Save(h.videos); err != nil {
		return ProgressState{}, fmt.Errorf("failed to save progress: %w", err)
	}

	current, next, found := video.FindNext(h.videos, path)
	return ProgressState{
		Current:      current,
		Next:         next,
		Found:        found,
		AutoPlay:     h.autoPlay,
		PlaybackRate: h.playbackRate,
	}, nil
}

func (h *AppHandler) toggleVideoCompletion(path string) (model.Video, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	var toggled model.Video
	for i, v := range h.videos {
		if v.Path == path {
			h.videos[i].Completed = !h.videos[i].Completed
			toggled = h.videos[i]
			break
		}
	}

	if err := h.progressStore.Save(h.videos); err != nil {
		return model.Video{}, fmt.Errorf("failed to save progress: %w", err)
	}
	return toggled, nil
}

func (h *AppHandler) renderPlayerResponse(w http.ResponseWriter, v model.Video, autoPlay bool, speed float64, shouldAutoplay bool) {
	data := struct {
		CurrentVideo   *model.Video
		AutoPlay       bool
		Speed          float64
		ShouldAutoplay bool
	}{
		CurrentVideo:   &v,
		AutoPlay:       autoPlay,
		Speed:          speed,
		ShouldAutoplay: shouldAutoplay,
	}

	if err := h.tmpl.ExecuteTemplate(w, "player", data); err != nil {
		http.Error(w, fmt.Sprintf("template rendering failed: %v", err), http.StatusInternalServerError)
	}
}


func (h *AppHandler) renderToggleOOB(w http.ResponseWriter, v model.Video) {
	var buf strings.Builder
	if err := h.tmpl.ExecuteTemplate(&buf, "toggle-btn", map[string]interface{}{"Video": v}); err != nil {
		http.Error(w, fmt.Sprintf("template rendering failed: %v", err), http.StatusInternalServerError)
		return
	}
	// Add hx-swap-oob="true" to the form tag for HTMX out-of-band swap
	html := strings.Replace(buf.String(), "<form ", "<form hx-swap-oob=\"true\" ", 1)
	w.Write([]byte(html))
}

func parseSpeed(s string) float64 {
	var speed float64
	if _, err := fmt.Sscanf(s, "%f", &speed); err != nil || speed <= 0 {
		return 1.0
	}
	return speed
}
