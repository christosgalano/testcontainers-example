package models

// Song represents a music song.
type Song struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Composer string `json:"composer"`
}
