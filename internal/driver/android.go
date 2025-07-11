package driver

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/kazuph/mcp-android-chrome/internal/loader"
	"github.com/kazuph/mcp-android-chrome/internal/platform"
)

// AndroidDriver implements Driver for Android devices using ADB
type AndroidDriver struct {
	config   AndroidConfig
	tabLoader *loader.HTTPTabLoader
}

// NewAndroidDriver creates a new Android driver
func NewAndroidDriver(config AndroidConfig) *AndroidDriver {
	return &AndroidDriver{
		config: config,
	}
}

// Start sets up ADB port forwarding
func (d *AndroidDriver) Start(ctx context.Context) error {
	if err := d.CheckEnvironment(); err != nil {
		return fmt.Errorf("environment check failed: %w", err)
	}

	// Check if Android device is connected
	if err := platform.CheckADBDeviceConnected(); err != nil {
		return fmt.Errorf("device connection check failed: %w", err)
	}

	// Setup ADB port forwarding using absolute path
	adbPath := platform.FindADBPath()
	cmd := exec.CommandContext(ctx, adbPath, "-d", "forward", 
		fmt.Sprintf("tcp:%d", d.config.Port),
		fmt.Sprintf("localabstract:%s", d.config.Socket))
	
	if d.config.Debug {
		fmt.Fprintf(os.Stderr, "Executing: %s\n", cmd.String())
	}
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to setup ADB port forwarding: %w", err)
	}

	// Wait for connection to be ready
	if d.config.Wait > 0 {
		time.Sleep(d.config.Wait)
	}

	// Initialize HTTP tab loader
	d.tabLoader = loader.NewHTTPTabLoader(d.GetURL(), d.config.Timeout, d.config.Debug)
	
	return nil
}

// Stop cleans up ADB port forwarding
func (d *AndroidDriver) Stop(ctx context.Context) error {
	if d.config.SkipCleanup {
		return nil
	}

	adbPath := platform.FindADBPath()
	cmd := exec.CommandContext(ctx, adbPath, "-d", "forward", "--remove",
		fmt.Sprintf("tcp:%d", d.config.Port))
	
	if d.config.Debug {
		fmt.Fprintf(os.Stderr, "Executing cleanup: %s\n", cmd.String())
	}
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to cleanup ADB port forwarding: %w", err)
	}
	
	return nil
}

// GetURL returns the Chrome DevTools Protocol URL
func (d *AndroidDriver) GetURL() string {
	return fmt.Sprintf("http://localhost:%d/json/list", d.config.Port)
}

// CheckEnvironment verifies ADB is available
func (d *AndroidDriver) CheckEnvironment() error {
	return platform.CheckADBAvailable()
}

// LoadTabs retrieves tabs from the Android device
func (d *AndroidDriver) LoadTabs(ctx context.Context) ([]loader.Tab, error) {
	if d.tabLoader == nil {
		return nil, fmt.Errorf("driver not started")
	}
	
	return d.tabLoader.LoadTabs(ctx)
}

// RestoreTabs implements RestoreDriver interface for Android
func (d *AndroidDriver) RestoreTabs(ctx context.Context, tabs []loader.Tab) error {
	if d.tabLoader == nil {
		return fmt.Errorf("driver not started")
	}
	
	baseURL := fmt.Sprintf("http://localhost:%d", d.config.Port)
	restorer := loader.NewHTTPTabRestorer(baseURL, d.config.Timeout, d.config.Debug)
	
	return restorer.RestoreTabs(ctx, tabs)
}

// CloseTab closes a single tab by its ID
func (d *AndroidDriver) CloseTab(ctx context.Context, tabID string) error {
	if d.tabLoader == nil {
		return fmt.Errorf("driver not started")
	}
	
	// First, verify the tab exists
	if exists, err := d.tabExists(ctx, tabID); err != nil {
		return fmt.Errorf("failed to verify tab existence: %w", err)
	} else if !exists {
		return fmt.Errorf("tab with ID '%s' does not exist", tabID)
	}
	
	closeURL := fmt.Sprintf("http://localhost:%d/json/close/%s", d.config.Port, tabID)
	
	if d.config.Debug {
		fmt.Fprintf(os.Stderr, "Closing tab: %s -> %s\n", tabID, closeURL)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", closeURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create close request: %w", err)
	}
	
	client := &http.Client{Timeout: d.config.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to close tab: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code when closing tab: %d", resp.StatusCode)
	}
	
	if d.config.Debug {
		fmt.Fprintf(os.Stderr, "Successfully closed tab: %s\n", tabID)
	}
	
	return nil
}

// tabExists checks if a tab with the given ID exists
func (d *AndroidDriver) tabExists(ctx context.Context, tabID string) (bool, error) {
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

// TabCloseResult represents the result of closing multiple tabs
type TabCloseResult struct {
	SuccessCount int
	FailedCount  int
	FailedTabIDs []string
	FailedErrors map[string]string // tabID -> error message
}

// CloseTabs closes multiple tabs by their IDs and returns detailed results
func (d *AndroidDriver) CloseTabs(ctx context.Context, tabIDs []string) error {
	if d.tabLoader == nil {
		return fmt.Errorf("driver not started")
	}
	
	if d.config.Debug {
		fmt.Fprintf(os.Stderr, "Closing %d tabs\n", len(tabIDs))
	}
	
	result := TabCloseResult{
		FailedTabIDs: make([]string, 0),
		FailedErrors: make(map[string]string),
	}
	
	for _, tabID := range tabIDs {
		if err := d.CloseTab(ctx, tabID); err != nil {
			if d.config.Debug {
				fmt.Fprintf(os.Stderr, "Failed to close tab %s: %v\n", tabID, err)
			}
			result.FailedCount++
			result.FailedTabIDs = append(result.FailedTabIDs, tabID)
			result.FailedErrors[tabID] = err.Error()
		} else {
			result.SuccessCount++
		}
	}
	
	if result.FailedCount > 0 {
		return fmt.Errorf("partially successful: closed %d/%d tabs successfully. Failed tabs: %v", 
			result.SuccessCount, len(tabIDs), result.FailedTabIDs)
	}
	
	if d.config.Debug {
		fmt.Fprintf(os.Stderr, "Successfully closed all %d tabs\n", len(tabIDs))
	}
	
	return nil
}