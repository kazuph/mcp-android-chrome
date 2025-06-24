package driver

import (
	"context"
	"fmt"
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

	// Setup ADB port forwarding
	cmd := exec.CommandContext(ctx, "adb", "-d", "forward", 
		fmt.Sprintf("tcp:%d", d.config.Port),
		fmt.Sprintf("localabstract:%s", d.config.Socket))
	
	if d.config.Debug {
		fmt.Printf("Executing: %s\n", cmd.String())
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

	cmd := exec.CommandContext(ctx, "adb", "-d", "forward", "--remove",
		fmt.Sprintf("tcp:%d", d.config.Port))
	
	if d.config.Debug {
		fmt.Printf("Executing cleanup: %s\n", cmd.String())
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