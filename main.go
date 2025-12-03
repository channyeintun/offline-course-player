package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type Video struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	Completed bool   `json:"completed"`
}

var videos []Video
var dataFile = "progress.json"

func main() {
	loadVideos()
	loadProgress()

	http.HandleFunc("/", handleHome)
	http.HandleFunc("/toggle", handleToggle)
	http.Handle("/videos/", http.StripPrefix("/videos/", http.FileServer(http.Dir("videos"))))

	fmt.Println("Server is running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func loadVideos() {
	videos = []Video{}
	err := filepath.Walk("videos", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".mp4") {

			relPath, err := filepath.Rel("videos", path)
			if err != nil {
				return err
			}

			relPath = filepath.ToSlash(relPath)
			videos = append(videos, Video{
				Name:      info.Name(),
				Path:      relPath,
				Completed: false,
			})
		}
		return nil
	})
	if err != nil {
		fmt.Println("Error walking through directory:", err)
	}
}

func loadProgress() {
	data, err := os.ReadFile(dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		fmt.Println("Error reading progress file:", err)
		return
	}

	var savedVideos []Video
	err = json.Unmarshal(data, &savedVideos)
	if err != nil {
		fmt.Println("Error unmarshaling progress data:", err)
		return
	}

	for i, video := range videos {
		for _, savedVideo := range savedVideos {
			if video.Path == savedVideo.Path {
				videos[i].Completed = savedVideo.Completed
				break
			}
		}
	}
}

func saveProgress() {
	data, err := json.Marshal(videos)
	if err != nil {
		fmt.Println("Error marshaling progress data:", err)
		return
	}

	err = os.WriteFile(dataFile, data, 0644)
	if err != nil {
		fmt.Println("Error writing progress file:", err)
	}
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	videosJSON, err := json.Marshal(videos)
	if err != nil {
		http.Error(w, "Failed to encode videos", http.StatusInternalServerError)
		return
	}
	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Offline Learning Platform</title>
	<link rel="stylesheet" href="https://unpkg.com/open-props"/>
	<link rel="stylesheet" href="https://unpkg.com/open-props/normalize.min.css"/>
	<style>
		:root {
			--sidebar-width: 400px;
			--accent-color: var(--indigo-6);
			--accent-hover: var(--indigo-7);
			--bg-color: var(--gray-0);
			--sidebar-bg: var(--gray-0);
			--text-main: var(--gray-9);
			--text-muted: var(--gray-6);
			--border-color: var(--gray-3);
		}

		body {
			display: grid;
			grid-template-columns: 1fr var(--sidebar-width);
			height: 100vh;
			overflow: hidden;
			font-family: var(--font-sans);
			background-color: var(--bg-color);
			color: var(--text-main);
		}

		/* Main Player Area */
		#video-player {
			display: flex;
			flex-direction: column;
			justify-content: center;
			align-items: center;
			background-color: #000; /* Cinematic background */
			padding: var(--size-5);
			position: relative;
		}

		#player {
			width: 100%;
			max-width: 1400px;
			aspect-ratio: 16/9;
			border-radius: var(--radius-3);
			box-shadow: var(--shadow-6);
			background-color: #000;
			outline: none;
		}

		.player-controls {
			margin-top: var(--size-4);
			display: flex;
			gap: var(--size-3);
			width: 100%;
			max-width: 1400px;
			justify-content: space-between;
			align-items: center;
			color: white; /* Since background is black */
		}

		#autoplay-btn {
			background: rgba(255, 255, 255, 0.1);
			color: var(--gray-1);
			border: 1px solid rgba(255, 255, 255, 0.2);
			padding: var(--size-2) var(--size-4);
			border-radius: var(--radius-pill);
			font-size: var(--font-size-1);
			font-weight: var(--font-weight-6);
			cursor: pointer;
			transition: all 0.2s ease;
			display: flex;
			align-items: center;
			gap: var(--size-2);
		}

		#autoplay-btn:hover {
			background: rgba(255, 255, 255, 0.2);
			transform: translateY(-1px);
		}

		#autoplay-btn.active {
			background: var(--accent-color);
			border-color: var(--accent-color);
			color: white;
		}

		/* Sidebar / Playlist */
		#video-list {
			background-color: var(--sidebar-bg);
			border-left: 1px solid var(--border-color);
			overflow-y: auto;
			display: flex;
			flex-direction: column;
		}

		.playlist-header {
			padding: var(--size-4);
			border-bottom: 1px solid var(--border-color);
			background: var(--sidebar-bg);
			position: sticky;
			top: 0;
			z-index: 10;
			backdrop-filter: blur(10px); /* If we make bg transparent later */
		}

		.playlist-header h2 {
			font-size: var(--font-size-4);
			font-weight: var(--font-weight-7);
			color: var(--text-main);
			margin: 0;
		}

		.playlist-content {
			padding: var(--size-3);
		}

		/* Sections */
		.section {
			margin-bottom: var(--size-3);
		}

		.section-title {
			padding: var(--size-2) var(--size-3);
			font-size: var(--font-size-1);
			font-weight: var(--font-weight-7);
			color: var(--text-muted);
			text-transform: uppercase;
			letter-spacing: var(--font-letterspacing-1);
			cursor: pointer;
			display: flex;
			justify-content: space-between;
			align-items: center;
			user-select: none;
			border-radius: var(--radius-2);
			transition: background-color 0.2s ease;
		}

		.section-title:hover {
			background-color: var(--gray-2);
			color: var(--text-main);
		}
		
		.section-title::after {
			content: '▼';
			font-size: 0.7em;
			transition: transform 0.2s ease;
		}
		
		.section.collapsed .section-title::after {
			transform: rotate(-90deg);
		}

		.section-content {
			margin-top: var(--size-1);
			display: flex;
			flex-direction: column;
			gap: 2px;
		}

		/* Video Items */
		.video-item {
			display: flex;
			align-items: center;
			gap: var(--size-3);
			padding: var(--size-2) var(--size-3);
			border-radius: var(--radius-2);
			cursor: pointer;
			transition: all 0.2s ease;
			text-decoration: none;
			color: var(--text-main);
			border: 1px solid transparent;
		}

		.video-item:hover {
			background-color: var(--gray-2);
		}

		.video-item.active {
			background-color: var(--indigo-0);
			border-color: var(--indigo-2);
			color: var(--indigo-9);
		}

		.video-item.active .video-title {
			font-weight: var(--font-weight-6);
		}

		/* Checkbox styling */
		.video-checkbox {
			appearance: none;
			width: 18px;
			height: 18px;
			border: 2px solid var(--gray-4);
			border-radius: 4px;
			cursor: pointer;
			position: relative;
			flex-shrink: 0;
			transition: all 0.2s ease;
		}

		.video-checkbox:checked {
			background-color: var(--green-6);
			border-color: var(--green-6);
		}

		.video-checkbox:checked::after {
			content: '✓';
			position: absolute;
			color: white;
			font-size: 12px;
			top: 50%;
			left: 50%;
			transform: translate(-50%, -50%);
			font-weight: bold;
		}

		.video-title {
			font-size: var(--font-size-1);
			line-height: 1.4;
			flex: 1;
		}

		/* Scrollbar */
		#video-list::-webkit-scrollbar {
			width: 8px;
		}
		#video-list::-webkit-scrollbar-track {
			background: transparent;
		}
		#video-list::-webkit-scrollbar-thumb {
			background-color: var(--gray-4);
			border-radius: 20px;
			border: 3px solid var(--sidebar-bg);
		}
		#video-list::-webkit-scrollbar-thumb:hover {
			background-color: var(--gray-5);
		}
		
		/* Responsive */
		@media (max-width: 900px) {
			body {
				grid-template-columns: 1fr;
				grid-template-rows: auto 1fr;
				overflow: auto;
			}
			#video-player {
				position: sticky;
				top: 0;
				z-index: 100;
				padding: 0;
			}
			#player {
				border-radius: 0;
			}
			#video-list {
				border-left: none;
				border-top: 1px solid var(--border-color);
				overflow: visible;
			}
			.playlist-header {
				position: static;
			}
		}
	</style>
</head>
<body>
	<div id="video-player">
		<video id="player" controls>
			<source src="" type="video/mp4">
			Your browser does not support the video tag.
		</video>
		<div class="player-controls">
			<div id="current-video-title" style="font-weight: 600; font-size: var(--font-size-2);">Select a video</div>
			<button id="autoplay-btn">
				<span>Auto Play</span>
				<span id="autoplay-status">Off</span>
			</button>
		</div>
	</div>

	<div id="video-list">
		<div class="playlist-header">
			<h2>Course Content</h2>
		</div>
		<div id="playlist-container" class="playlist-content">
			<!-- Video items will be injected here -->
		</div>
	</div>

	<script>
		const videos = {{.VideosJSON}};
		let currentPlayingPath = null;
		let autoPlay = false;

		// Initialize
		document.addEventListener('DOMContentLoaded', () => {
			renderVideos(videos);
			setupAutoPlay();
		});

		function setupAutoPlay() {
			const btn = document.getElementById('autoplay-btn');
			const status = document.getElementById('autoplay-status');
			
			btn.onclick = () => {
				autoPlay = !autoPlay;
				status.textContent = autoPlay ? "On" : "Off";
				btn.classList.toggle('active', autoPlay);
			};
		}

		function playVideo(path, name) {
			const player = document.getElementById('player');
			const titleDisplay = document.getElementById('current-video-title');
			
			player.src = '/videos/' + encodeURIComponent(path);
			currentPlayingPath = path;
			titleDisplay.textContent = name.replace(/\.[^/.]+$/, ""); // Remove extension
			
			renderVideos(videos);
			player.play();

			player.onended = () => {
				toggleCompleted(path, true); // Mark as completed locally and on server
				if (autoPlay) {
					playNextVideo(path);
				}
			};
		}

		function playNextVideo(currentPath) {
			const flatList = getFlatVideoList();
			const currentIndex = flatList.findIndex(v => v.path === currentPath);
			
			if (currentIndex !== -1 && currentIndex < flatList.length - 1) {
				const nextVideo = flatList[currentIndex + 1];
				playVideo(nextVideo.path, nextVideo.name);
			}
		}

		function getFlatVideoList() {
			// Helper to get a flat list of videos respecting the section order
			const sections = groupVideosBySection(videos);
			const flatList = [];
			
			// If we have sections
			if (Object.keys(sections).length > 0) {
				// Sort sections if needed, or rely on object insertion order (usually fine for simple cases)
				// Assuming sections come in a reasonable order or we just iterate keys
				for (const sectionName in sections) {
					flatList.push(...sections[sectionName]);
				}
			} else {
				flatList.push(...videos);
			}
			return flatList;
		}

		function groupVideosBySection(videos) {
			const sections = {};
			videos.forEach(video => {
				const parts = video.path.split('/');
				if (parts.length === 2) {
					const sectionTitle = parts[0];
					if (!sections[sectionTitle]) sections[sectionTitle] = [];
					sections[sectionTitle].push(video);
				}
			});
			return sections;
		}

		function toggleCompleted(path, forceState = null) {
			// Update local state
			const video = videos.find(v => v.path === path);
			if (video) {
				video.completed = forceState !== null ? forceState : !video.completed;
				renderVideos(videos); // Re-render to show checkmark
			}

			// Update server state
			fetch('/toggle', {
				method: 'POST',
				headers: {'Content-Type': 'application/x-www-form-urlencoded'},
				body: 'path=' + encodeURIComponent(path)
			}).catch(console.error);
		}

		function renderVideos(videoData) {
			const container = document.getElementById('playlist-container');
			container.innerHTML = '';

			const sections = groupVideosBySection(videoData);
			const hasSections = Object.keys(sections).length > 0;

			if (hasSections) {
				for (const [sectionName, sectionVideos] of Object.entries(sections)) {
					const sectionDiv = document.createElement('div');
					sectionDiv.className = 'section';
					
					// Check if this section contains the current video to auto-expand
					const containsActive = sectionVideos.some(v => v.path === currentPlayingPath);
					
					const title = document.createElement('div');
					title.className = 'section-title';
					title.textContent = sectionName.replace(/^\d+[_-]/, '').replace(/_/g, ' '); // Clean up title
					title.onclick = () => {
						sectionDiv.classList.toggle('collapsed');
						const content = sectionDiv.querySelector('.section-content');
						content.style.display = content.style.display === 'none' ? 'flex' : 'none';
					};
					
					const content = document.createElement('div');
					content.className = 'section-content';
					// If not active and not the first section, maybe collapse? Let's keep all open by default for now
					// or collapse if we want to save space. Let's keep open.
					
					sectionVideos.forEach(video => {
						content.appendChild(createVideoItem(video));
					});

					sectionDiv.appendChild(title);
					sectionDiv.appendChild(content);
					container.appendChild(sectionDiv);
				}
			} else {
				// No sections, just list
				videoData.forEach(video => {
					container.appendChild(createVideoItem(video));
				});
			}
		}

		function createVideoItem(video) {
			const item = document.createElement('div');
			item.className = 'video-item';
			if (video.path === currentPlayingPath) item.classList.add('active');

			const checkbox = document.createElement('input');
			checkbox.type = 'checkbox';
			checkbox.className = 'video-checkbox';
			checkbox.checked = video.completed;
			checkbox.onclick = (e) => {
				e.stopPropagation();
				toggleCompleted(video.path);
			};

			const title = document.createElement('span');
			title.className = 'video-title';
			title.textContent = video.name.replace(/^\d+[_-]/, '').replace(/\.mp4$/i, '').replace(/_/g, ' ');
			
			item.onclick = () => playVideo(video.path, video.name);

			item.appendChild(checkbox);
			item.appendChild(title);
			
			return item;
		}
	</script>
</body>
</html>`

	t, err := template.New("home").Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, struct {
		VideosJSON template.JS
	}{
		VideosJSON: template.JS(videosJSON),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleToggle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := r.FormValue("path")
	decodedPath, err := url.QueryUnescape(path) // Decode the URL-encoded path
	if err != nil {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	for i, video := range videos {
		if video.Path == decodedPath {
			videos[i].Completed = !videos[i].Completed
			break
		}
	}

	saveProgress()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}
