package loader

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"
)

// HTTPTabLoader handles HTTP-based tab loading via Chrome DevTools Protocol
type HTTPTabLoader struct {
	url     string
	timeout time.Duration
	debug   bool
	client  *http.Client
}

// NewHTTPTabLoader creates a new HTTP tab loader
func NewHTTPTabLoader(url string, timeout time.Duration, debug bool) *HTTPTabLoader {
	return &HTTPTabLoader{
		url:     url,
		timeout: timeout,
		debug:   debug,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// LoadTabs retrieves tabs from Chrome DevTools Protocol endpoint
func (h *HTTPTabLoader) LoadTabs(ctx context.Context) ([]Tab, error) {
	if h.debug {
		fmt.Fprintf(os.Stderr, "Loading tabs from: %s\n", h.url)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", h.url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tabs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var tabs []Tab
	if err := json.NewDecoder(resp.Body).Decode(&tabs); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if h.debug {
		fmt.Fprintf(os.Stderr, "Loaded %d tabs\n", len(tabs))
	}

	return tabs, nil
}

// HTTPTabRestorer handles HTTP-based tab restoration
type HTTPTabRestorer struct {
	baseURL string
	timeout time.Duration
	debug   bool
	client  *http.Client
}

// NewHTTPTabRestorer creates a new HTTP tab restorer
func NewHTTPTabRestorer(baseURL string, timeout time.Duration, debug bool) *HTTPTabRestorer {
	return &HTTPTabRestorer{
		baseURL: baseURL,
		timeout: timeout,
		debug:   debug,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// RestoreTabs restores tabs using Chrome DevTools Protocol
func (h *HTTPTabRestorer) RestoreTabs(ctx context.Context, tabs []Tab) error {
	if h.debug {
		fmt.Fprintf(os.Stderr, "Restoring %d tabs\n", len(tabs))
	}

	for i, tab := range tabs {
		if err := h.restoreTab(ctx, tab, i); err != nil {
			return fmt.Errorf("failed to restore tab %d (%s): %w", i, tab.Title, err)
		}
	}

	return nil
}

// restoreTab restores a single tab
func (h *HTTPTabRestorer) restoreTab(ctx context.Context, tab Tab, index int) error {
	// Construct URL for creating new tab
	createURL := fmt.Sprintf("%s/json/new?%s", h.baseURL, url.QueryEscape(tab.URL))
	
	if h.debug {
		fmt.Fprintf(os.Stderr, "Restoring tab %d: %s -> %s\n", index+1, tab.Title, tab.URL)
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", createURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to restore tab: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Small delay between tab restorations to avoid overwhelming the browser
	time.Sleep(100 * time.Millisecond)

	return nil
}