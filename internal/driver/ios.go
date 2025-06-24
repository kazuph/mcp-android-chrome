package driver

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/kazuph/mcp-android-chrome/internal/loader"
	"github.com/kazuph/mcp-android-chrome/internal/platform"
)

// IOSDriver implements Driver for iOS devices using iOS WebKit Debug Proxy
type IOSDriver struct {
	config   IOSConfig
	cmd      *exec.Cmd
	tabLoader *loader.HTTPTabLoader
}

// NewIOSDriver creates a new iOS driver
func NewIOSDriver(config IOSConfig) *IOSDriver {
	return &IOSDriver{
		config: config,
	}
}

// Start launches ios_webkit_debug_proxy as a background process
func (d *IOSDriver) Start(ctx context.Context) error {
	if err := d.CheckEnvironment(); err != nil {
		return fmt.Errorf("environment check failed: %w", err)
	}

	// Start ios_webkit_debug_proxy
	args := []string{"-F", "-c", "null:9221,:9222-9322"}
	if d.config.Debug {
		args = append(args, "--debug")
	}
	
	proxyPath := platform.FindIOSWebKitDebugProxyPath()
	d.cmd = exec.CommandContext(ctx, proxyPath, args...)
	
	if d.config.Debug {
		fmt.Fprintf(os.Stderr, "Executing: %s\n", d.cmd.String())
	}
	
	if err := d.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ios_webkit_debug_proxy: %w", err)
	}

	// Wait for proxy to be ready
	if d.config.Wait > 0 {
		time.Sleep(d.config.Wait)
	}

	// Initialize HTTP tab loader
	d.tabLoader = loader.NewHTTPTabLoader(d.GetURL(), d.config.Timeout, d.config.Debug)
	
	return nil
}

// Stop terminates the ios_webkit_debug_proxy process
func (d *IOSDriver) Stop(ctx context.Context) error {
	if d.cmd == nil {
		return nil
	}

	if d.config.Debug {
		fmt.Fprintln(os.Stderr, "Terminating ios_webkit_debug_proxy process")
	}

	if err := d.cmd.Process.Kill(); err != nil {
		return fmt.Errorf("failed to kill ios_webkit_debug_proxy: %w", err)
	}

	// Wait for process to finish
	_ = d.cmd.Wait()
	d.cmd = nil
	
	return nil
}

// GetURL returns the WebKit Debug Proxy URL
func (d *IOSDriver) GetURL() string {
	return fmt.Sprintf("http://localhost:%d/json", d.config.Port)
}

// CheckEnvironment verifies ios_webkit_debug_proxy is available
func (d *IOSDriver) CheckEnvironment() error {
	return platform.CheckIOSWebKitDebugProxyAvailable()
}

// LoadTabs retrieves tabs from the iOS device
func (d *IOSDriver) LoadTabs(ctx context.Context) ([]loader.Tab, error) {
	if d.tabLoader == nil {
		return nil, fmt.Errorf("driver not started")
	}
	
	return d.tabLoader.LoadTabs(ctx)
}

// RestoreTabs implements RestoreDriver interface for iOS using WebSocket
func (d *IOSDriver) RestoreTabs(ctx context.Context, tabs []loader.Tab) error {
	if d.cmd == nil {
		return fmt.Errorf("driver not started")
	}

	// For iOS restoration, we need to use the WebSocket approach
	// This is more complex and requires creating an HTML file with WebSocket client
	baseURL := fmt.Sprintf("http://localhost:%d", d.config.Port)
	restorer := loader.NewWebSocketTabRestorer(baseURL, d.config.Debug)
	
	return restorer.RestoreTabs(ctx, tabs)
}

// CloseTab closes a single tab by its ID (iOS implementation)
func (d *IOSDriver) CloseTab(ctx context.Context, tabID string) error {
	if d.tabLoader == nil {
		return fmt.Errorf("driver not started")
	}
	
	// First, verify the tab exists
	if exists, err := d.tabExists(ctx, tabID); err != nil {
		return fmt.Errorf("failed to verify tab existence: %w", err)
	} else if !exists {
		return fmt.Errorf("tab with ID '%s' does not exist", tabID)
	}
	
	// iOS tab closing via WebSocket message
	return d.closeTabViaWebSocket(ctx, tabID)
}

// CloseTabs closes multiple tabs by their IDs (iOS implementation)
func (d *IOSDriver) CloseTabs(ctx context.Context, tabIDs []string) error {
	if d.tabLoader == nil {
		return fmt.Errorf("driver not started")
	}
	
	if d.config.Debug {
		fmt.Fprintf(os.Stderr, "Closing %d tabs on iOS\n", len(tabIDs))
	}
	
	successCount := 0
	var failedTabs []string
	
	for _, tabID := range tabIDs {
		if err := d.CloseTab(ctx, tabID); err != nil {
			if d.config.Debug {
				fmt.Fprintf(os.Stderr, "Failed to close iOS tab %s: %v\n", tabID, err)
			}
			failedTabs = append(failedTabs, tabID)
		} else {
			successCount++
		}
	}
	
	if len(failedTabs) > 0 {
		return fmt.Errorf("partially successful: closed %d/%d tabs successfully. Failed tabs: %v", 
			successCount, len(tabIDs), failedTabs)
	}
	
	if d.config.Debug {
		fmt.Fprintf(os.Stderr, "Successfully closed all %d iOS tabs\n", len(tabIDs))
	}
	
	return nil
}

// tabExists checks if a tab with the given ID exists (iOS)
func (d *IOSDriver) tabExists(ctx context.Context, tabID string) (bool, error) {
	tabs, err := d.LoadTabs(ctx)
	if err != nil {
		return false, err
	}
	
	for _, tab := range tabs {
		if tab.ID == tabID {
			return true, nil
		}
	}
	
	return false, nil
}

// closeTabViaWebSocket closes a tab using WebSocket communication
func (d *IOSDriver) closeTabViaWebSocket(ctx context.Context, tabID string) error {
	// For iOS, we use the simpler approach of sending JavaScript to close the tab
	// This is more reliable than complex WebSocket protocol implementations
	
	if d.config.Debug {
		fmt.Fprintf(os.Stderr, "Closing iOS tab %s via WebSocket\n", tabID)
	}
	
	// Create a WebSocket restorer to send close command
	baseURL := fmt.Sprintf("http://localhost:%d", d.config.Port)
	restorer := loader.NewWebSocketTabRestorer(baseURL, d.config.Debug)
	
	// Create a "fake" tab with JavaScript to close the window
	closeTabs := []loader.Tab{
		{
			ID:    tabID,
			Title: "Close Tab Command",
			URL:   "javascript:window.close()",
		},
	}
	
	// Execute the close command
	if err := restorer.RestoreTabs(ctx, closeTabs); err != nil {
		return fmt.Errorf("failed to send close command to iOS tab: %w", err)
	}
	
	return nil
}