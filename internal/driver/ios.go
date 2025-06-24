package driver

import (
	"context"
	"fmt"
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
		fmt.Printf("Executing: %s\n", d.cmd.String())
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
		fmt.Println("Terminating ios_webkit_debug_proxy process")
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