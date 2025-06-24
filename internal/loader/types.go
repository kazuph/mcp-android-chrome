package loader

// Tab represents a browser tab
type Tab struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	URL   string `json:"url"`
	Type  string `json:"type,omitempty"`
}