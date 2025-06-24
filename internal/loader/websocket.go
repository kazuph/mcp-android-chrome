package loader

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"path/filepath"
	"time"

	"github.com/gorilla/websocket"
	"github.com/kazuph/mcp-android-chrome/internal/platform"
	"github.com/kazuph/mcp-android-chrome/internal/template"
)

// WebSocketTabRestorer handles WebSocket-based tab restoration for iOS
type WebSocketTabRestorer struct {
	baseURL string
	debug   bool
}

// NewWebSocketTabRestorer creates a new WebSocket tab restorer
func NewWebSocketTabRestorer(baseURL string, debug bool) *WebSocketTabRestorer {
	return &WebSocketTabRestorer{
		baseURL: baseURL,
		debug:   debug,
	}
}

// RestoreTabs restores tabs using WebKit Debug Protocol via WebSocket
func (w *WebSocketTabRestorer) RestoreTabs(ctx context.Context, tabs []Tab) error {
	if w.debug {
		fmt.Printf("Restoring %d tabs via WebSocket\n", len(tabs))
	}

	// First, get the target page ID
	targetPageID, err := w.getTargetPageID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get target page ID: %w", err)
	}

	// Create temporary HTML file with WebSocket client
	htmlFile, err := w.createWebSocketClient(tabs, targetPageID)
	if err != nil {
		return fmt.Errorf("failed to create WebSocket client: %w", err)
	}

	// Open the HTML file in browser to execute the restoration
	if err := platform.OpenInBrowser("file://" + htmlFile); err != nil {
		return fmt.Errorf("failed to open browser: %w", err)
	}

	fmt.Printf("WebSocket client opened in browser. Please check your iOS device for restored tabs.\n")
	fmt.Printf("HTML file: %s\n", htmlFile)

	return nil
}

// getTargetPageID retrieves the target page ID for WebSocket communication
func (w *WebSocketTabRestorer) getTargetPageID(ctx context.Context) (string, error) {
	// Make HTTP request to get available targets
	loader := NewHTTPTabLoader(w.baseURL+"/json", 10*time.Second, w.debug)
	targets, err := loader.LoadTabs(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to load targets: %w", err)
	}

	// Find a suitable target (usually the first page)
	for _, target := range targets {
		if target.Type == "page" || target.Type == "" {
			return target.ID, nil
		}
	}

	return "", fmt.Errorf("no suitable target page found")
}

// createWebSocketClient creates an HTML file with embedded WebSocket client
func (w *WebSocketTabRestorer) createWebSocketClient(tabs []Tab, targetPageID string) (string, error) {
	// Parse base URL to get WebSocket URL
	u, err := url.Parse(w.baseURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse base URL: %w", err)
	}

	wsURL := fmt.Sprintf("ws://%s/devtools/page/%s", u.Host, targetPageID)

	// Convert tabs to JSON for JavaScript
	tabsJSON, err := json.Marshal(tabs)
	if err != nil {
		return "", fmt.Errorf("failed to marshal tabs: %w", err)
	}

	// Create HTML content using template
	htmlGenerator := template.NewWebSocketClientTemplate()
	htmlContent := htmlGenerator.Generate(string(tabsJSON), wsURL, w.debug)

	// Write to temporary file
	filename := fmt.Sprintf("tab-restore-%d.html", time.Now().Unix())
	filepath := filepath.Join("/tmp", filename)
	
	if err := template.WriteFile(filepath, htmlContent); err != nil {
		return "", fmt.Errorf("failed to write HTML file: %w", err)
	}

	return filepath, nil
}

// WebSocketMessage represents a WebKit Debug Protocol message
type WebSocketMessage struct {
	ID     int                    `json:"id"`
	Method string                 `json:"method"`
	Params map[string]interface{} `json:"params"`
}

// Alternative direct WebSocket implementation (if needed)
func (w *WebSocketTabRestorer) restoreTabsDirect(ctx context.Context, tabs []Tab, targetPageID string) error {
	u, err := url.Parse(w.baseURL)
	if err != nil {
		return fmt.Errorf("failed to parse base URL: %w", err)
	}

	wsURL := fmt.Sprintf("ws://%s/devtools/page/%s", u.Host, targetPageID)
	
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, wsURL, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to WebSocket: %w", err)
	}
	defer conn.Close()

	for i, tab := range tabs {
		msg := WebSocketMessage{
			ID:     i + 1,
			Method: "Target.sendMessageToTarget",
			Params: map[string]interface{}{
				"targetId": targetPageID,
				"message": fmt.Sprintf(`{"id":1,"method":"Runtime.evaluate","params":{"expression":"window.open('%s');"}}`, tab.URL),
			},
		}

		if err := conn.WriteJSON(msg); err != nil {
			return fmt.Errorf("failed to send WebSocket message: %w", err)
		}

		// Read response
		var response map[string]interface{}
		if err := conn.ReadJSON(&response); err != nil {
			return fmt.Errorf("failed to read WebSocket response: %w", err)
		}

		if w.debug {
			fmt.Printf("Restored tab %d: %s\n", i+1, tab.Title)
		}

		// Small delay between restorations
		time.Sleep(200 * time.Millisecond)
	}

	return nil
}