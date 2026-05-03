package model

// Video represents a single course video.
type Video struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	Completed bool   `json:"completed"`
}

// Section logically groups a list of videos together.
type Section struct {
	Name   string
	Videos []Video
}
