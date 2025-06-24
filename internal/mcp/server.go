package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	mcp_golang "github.com/metoro-io/mcp-golang"
	"github.com/metoro-io/mcp-golang/transport/stdio"

	"github.com/kazuph/mcp-android-chrome/internal/driver"
	"github.com/kazuph/mcp-android-chrome/internal/format"
	"github.com/kazuph/mcp-android-chrome/internal/loader"
	"github.com/kazuph/mcp-android-chrome/internal/platform"
)

// TabTransferServer implements MCP server for tab transfer functionality
type TabTransferServer struct {
	server      *mcp_golang.Server
	tabCache    []loader.Tab
	cacheMutex  sync.RWMutex
	cacheSize   int
	lastUpdated time.Time
}

// NewTabTransferServer creates a new MCP server for tab transfer
func NewTabTransferServer() *TabTransferServer {
	server := mcp_golang.NewServer(stdio.NewStdioServerTransport())
	
	// Default cache size is 30, can be overridden by environment variable
	cacheSize := 30
	if envSize := os.Getenv("TAB_CACHE_SIZE"); envSize != "" {
		if size := parseInt(envSize); size > 0 {
			cacheSize = size
		}
	}
	
	return &TabTransferServer{
		server:    server,
		tabCache:  make([]loader.Tab, 0),
		cacheSize: cacheSize,
	}
}

// parseInt safely converts string to int
func parseInt(s string) int {
	if s == "" {
		return 0
	}
	result := 0
	for _, r := range s {
		if r >= '0' && r <= '9' {
			result = result*10 + int(r-'0')
		} else {
			return 0
		}
	}
	return result
}

// Start initializes and starts the MCP server
func (s *TabTransferServer) Start() error {
	// Register MCP tools
	if err := s.registerTools(); err != nil {
		return fmt.Errorf("failed to register tools: %w", err)
	}

	// Register MCP resources
	if err := s.registerResources(); err != nil {
		return fmt.Errorf("failed to register resources: %w", err)
	}

	// Auto-populate tab cache on startup (non-blocking)
	go s.populateTabCache()

	// Start the server
	err := s.server.Serve()
	if err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	// Keep the server running
	select {}
}

// populateTabCache attempts to fetch and cache Android tabs on startup
func (s *TabTransferServer) populateTabCache() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "Tab cache population failed with panic: %v\n", r)
		}
	}()

	// Try to populate cache with Android tabs
	if err := s.fetchAndCacheAndroidTabs(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to populate tab cache: %v\n", err)
		// Don't fail the server startup if cache population fails
	} else {
		fmt.Fprintf(os.Stderr, "Successfully populated tab cache with %d tabs\n", len(s.tabCache))
	}
}

// fetchAndCacheAndroidTabs fetches tabs from Android device and updates cache
func (s *TabTransferServer) fetchAndCacheAndroidTabs() error {
	config := driver.AndroidConfig{
		DriverConfig: driver.DriverConfig{
			Port:    9222,
			Timeout: 10 * time.Second,
			Debug:   false, // Don't spam logs during auto-fetch
		},
		Socket: "chrome_devtools_remote",
		Wait:   2 * time.Second,
	}

	androidDriver := driver.NewAndroidDriver(config)
	
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Start driver
	if err := androidDriver.Start(ctx); err != nil {
		return fmt.Errorf("failed to start Android driver: %w", err)
	}
	defer androidDriver.Stop(ctx)

	// Load tabs
	tabs, err := androidDriver.LoadTabs(ctx)
	if err != nil {
		return fmt.Errorf("failed to load tabs: %w", err)
	}

	// Update cache with latest tabs (limit to cacheSize)
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()
	
	if len(tabs) > s.cacheSize {
		s.tabCache = tabs[:s.cacheSize]
	} else {
		s.tabCache = tabs
	}
	s.lastUpdated = time.Now()

	return nil
}

// registerTools registers all available MCP tools
func (s *TabTransferServer) registerTools() error {
	// Tool 1: Copy tabs from Android
	err := s.server.RegisterTool("copy_tabs_android", `Copy Chrome tabs from Android device via ADB.

Prerequisites:
1. Android device with USB debugging enabled (Settings > Developer Options > USB Debugging)
2. ADB (Android Debug Bridge) installed and in PATH
3. Chrome browser running on Android device
4. USB cable connecting device to computer
5. Device unlocked and USB debugging permission granted

Common Issues & Solutions:
- "adb command not found": Install Android Platform Tools
  - macOS: brew install --cask android-platform-tools
  - Linux: sudo apt install android-tools-adb
  - Windows: Download from developer.android.com
- "device unauthorized": Check device screen for USB debugging prompt and tap "Allow"
- "no devices found": Ensure USB cable supports data transfer (not just charging)
- "connection refused": Restart ADB with 'adb kill-server && adb start-server'

This tool will automatically check environment and provide specific error messages if prerequisites are not met.`, s.copyTabsAndroid)
	if err != nil {
		return fmt.Errorf("failed to register copy_tabs_android: %w", err)
	}

	// Tool 2: Copy tabs from iOS
	err = s.server.RegisterTool("copy_tabs_ios", `Copy Chrome/Safari tabs from iOS device via WebKit Debug Proxy.

Prerequisites:
1. iOS device with Web Inspector enabled (Settings > Safari > Advanced > Web Inspector)
2. For Chrome: Enable Web Inspector in Chrome Settings > Privacy and Security > Site Settings
3. iOS WebKit Debug Proxy installed and in PATH
4. USB cable connecting device to computer
5. Trust computer when prompted on iOS device
6. iOS 16.4+ for Chrome support (Safari works on older versions)

Common Issues & Solutions:
- "ios_webkit_debug_proxy command not found": Install WebKit Debug Proxy
  - macOS: brew install ios-webkit-debug-proxy
  - Linux: See github.com/google/ios-webkit-debug-proxy for build instructions
  - Windows: Not officially supported
- "Could not connect to device": Ensure device is unlocked and trusted
- "No targets found": Make sure Safari/Chrome is running and has open tabs
- "Connection timeout": Try disconnecting and reconnecting USB cable

This tool will automatically check environment and provide specific error messages if prerequisites are not met.`, s.copyTabsIOS)
	if err != nil {
		return fmt.Errorf("failed to register copy_tabs_ios: %w", err)
	}

	// Tool 3: Reopen tabs
	err = s.server.RegisterTool("reopen_tabs", `Restore saved tabs to mobile device.

This tool takes previously exported tabs (from copy_tabs_android or copy_tabs_ios) and reopens them on the target device.

Prerequisites (same as copy tools):
- For Android: ADB installed, USB debugging enabled, device connected
- For iOS: iOS WebKit Debug Proxy installed, Web Inspector enabled, device connected

The tool automatically detects platform-specific requirements and provides detailed error messages for troubleshooting.`, s.reopenTabs)
	if err != nil {
		return fmt.Errorf("failed to register reopen_tabs: %w", err)
	}

	// Tool 4: Check environment
	err = s.server.RegisterTool("check_environment", `Check system dependencies and device connectivity.

This diagnostic tool verifies:
1. ADB (Android Debug Bridge) installation and functionality
2. iOS WebKit Debug Proxy installation and functionality
3. Device connectivity status
4. USB debugging permissions

Use this tool first to diagnose setup issues before attempting tab operations. It provides specific installation commands and troubleshooting steps for each platform.`, s.checkEnvironment)
	if err != nil {
		return fmt.Errorf("failed to register check_environment: %w", err)
	}

	// Tool 5: Refresh tab cache
	err = s.server.RegisterTool("refresh_tab_cache", `Manually refresh the current tab cache from Android device.

This tool fetches the latest tabs from the connected Android device and updates the internal cache. Useful when you want to ensure the current_tabs resource reflects the most recent browser state.

The cache is automatically populated on server startup, but this tool allows manual updates without restarting the server.`, s.refreshTabCache)
	if err != nil {
		return fmt.Errorf("failed to register refresh_tab_cache: %w", err)
	}

	// Tool 6: Cache status
	err = s.server.RegisterTool("cache_status", `Check the current status of the tab cache.

This diagnostic tool shows:
- Number of cached tabs
- Cache size limit
- Last update timestamp
- Cache population status

Useful for debugging cache-related issues and understanding the current state of cached data.`, s.cacheStatus)
	if err != nil {
		return fmt.Errorf("failed to register cache_status: %w", err)
	}

	// Tool 7: Close single tab
	err = s.server.RegisterTool("close_tab", `Close a single tab on Android device by tab ID.

This tool closes a specific tab using its unique Chrome DevTools Protocol ID. The tab ID can be obtained from copy_tabs_android tool or current_tabs resource.

⚠️ Warning: This action cannot be undone. The tab will be permanently closed.

Arguments:
- tabId (required): The unique ID of the tab to close
- platform (optional): Target platform (default: android)
- confirm (optional): Set to true to skip confirmation (default: false)

Safety: Use cache_status or copy_tabs_android first to get current tab IDs.`, s.closeTab)
	if err != nil {
		return fmt.Errorf("failed to register close_tab: %w", err)
	}

	// Tool 8: Close multiple tabs
	err = s.server.RegisterTool("close_tabs_bulk", `Close multiple tabs at once on Android device.

This tool allows bulk closing of tabs by their IDs or by filtering criteria. Useful for cleaning up many tabs simultaneously.

⚠️ Warning: This action cannot be undone. All matching tabs will be permanently closed.

Arguments:
- tabIds (optional): Array of specific tab IDs to close
- platform (optional): Target platform (default: android)
- filterUrl (optional): Close tabs matching URL pattern (supports wildcards)
- filterTitle (optional): Close tabs matching title pattern (supports wildcards)
- confirm (optional): Set to true to skip confirmation (default: false)
- dryRun (optional): Preview which tabs would be closed without actually closing them

Safety: Use dryRun=true first to preview the operation.`, s.closeTabsBulk)
	if err != nil {
		return fmt.Errorf("failed to register close_tabs_bulk: %w", err)
	}

	// Tool 9: Search tabs
	err = s.server.RegisterTool("search_tabs", `Search through currently cached tabs with advanced filtering and ranking.

This tool provides powerful search capabilities across cached tabs, including:
- Full-text search across URLs and titles
- Fuzzy matching for partial queries
- Domain-based filtering
- Relevance scoring and ranking
- Multiple search criteria combination

Arguments:
- query (optional): Search query to match against URLs and titles
- domain (optional): Filter by specific domain (e.g., "github.com")
- title (optional): Search specifically in tab titles
- url (optional): Search specifically in URLs
- limit (optional): Maximum number of results to return (default: 10)
- format (optional): Output format: json or yaml (default: json)

Returns ranked results with relevance scores for better search experience.`, s.searchTabs)
	if err != nil {
		return fmt.Errorf("failed to register search_tabs: %w", err)
	}

	return nil
}

// registerResources registers MCP resources
func (s *TabTransferServer) registerResources() error {
	// Resource: Current tabs in YAML format only
	err := s.server.RegisterResource("tabs://current", "current_tabs", "Currently loaded tabs (YAML format)", "application/x-yaml", s.getCurrentTabsYAML)
	if err != nil {
		return fmt.Errorf("failed to register current_tabs resource: %w", err)
	}

	return nil
}

// AndroidTabsArgs represents arguments for Android tab copying
type AndroidTabsArgs struct {
	Port        int    `json:"port" jsonschema:"description=Port for ADB forwarding (default: 9222)"`
	Socket      string `json:"socket" jsonschema:"description=ADB socket name (default: chrome_devtools_remote)"`
	Timeout     int    `json:"timeout" jsonschema:"description=Network timeout in seconds (default: 10)"`
	Wait        int    `json:"wait" jsonschema:"description=Wait time before starting in seconds (default: 2)"`
	SkipCleanup bool   `json:"skipCleanup" jsonschema:"description=Skip ADB cleanup after operation"`
	Debug       bool   `json:"debug" jsonschema:"description=Enable debug output"`
	Format      string `json:"format" jsonschema:"description=Output format: json or yaml (default: json)"`
}

// IOSTabsArgs represents arguments for iOS tab copying
type IOSTabsArgs struct {
	Port    int    `json:"port" jsonschema:"description=Port for iOS WebKit Debug Proxy (default: 9222)"`
	Timeout int    `json:"timeout" jsonschema:"description=Network timeout in seconds (default: 10)"`
	Wait    int    `json:"wait" jsonschema:"description=Wait time before starting in seconds (default: 2)"`
	Debug   bool   `json:"debug" jsonschema:"description=Enable debug output"`
	Format  string `json:"format" jsonschema:"description=Output format: json or yaml (default: json)"`
}

// ReopenTabsArgs represents arguments for tab restoration
type ReopenTabsArgs struct {
	TabsJSON    string `json:"tabsJson" jsonschema:"required,description=JSON string containing tabs to restore"`
	Platform    string `json:"platform" jsonschema:"required,description=Target platform (android or ios)"`
	Port        int    `json:"port" jsonschema:"description=Port for device communication (default: 9222)"`
	Timeout     int    `json:"timeout" jsonschema:"description=Network timeout in seconds (default: 10)"`
	Debug       bool   `json:"debug" jsonschema:"description=Enable debug output"`
}

// CheckEnvironmentArgs represents arguments for environment checking
type CheckEnvironmentArgs struct {
	Platform string `json:"platform" jsonschema:"description=Platform: android, ios, or all"`
}

// copyTabsAndroid implements the Android tab copying tool
func (s *TabTransferServer) copyTabsAndroid(args AndroidTabsArgs) (*mcp_golang.ToolResponse, error) {
	// Set defaults
	if args.Port == 0 {
		args.Port = 9222
	}
	if args.Socket == "" {
		args.Socket = "chrome_devtools_remote"
	}
	if args.Timeout == 0 {
		args.Timeout = 10
	}
	if args.Wait == 0 {
		args.Wait = 2
	}

	config := driver.AndroidConfig{
		DriverConfig: driver.DriverConfig{
			Port:    args.Port,
			Timeout: time.Duration(args.Timeout) * time.Second,
			Debug:   args.Debug,
		},
		Socket:      args.Socket,
		Wait:        time.Duration(args.Wait) * time.Second,
		SkipCleanup: args.SkipCleanup,
	}

	androidDriver := driver.NewAndroidDriver(config)
	
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(args.Timeout+10)*time.Second)
	defer cancel()

	// Start driver
	if err := androidDriver.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start Android driver: %w", err)
	}
	defer androidDriver.Stop(ctx)

	// Load tabs
	tabs, err := androidDriver.LoadTabs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load tabs: %w", err)
	}

	// Determine output format
	outputFormat := format.FormatJSON
	if args.Format != "" {
		if parsedFormat, err := format.ParseFormat(args.Format); err == nil {
			outputFormat = parsedFormat
		}
	}

	// Format tabs according to specified format
	formatter := format.NewTabFormatter(outputFormat)
	formattedTabs, err := formatter.FormatTabs(tabs)
	if err != nil {
		return nil, fmt.Errorf("failed to format tabs: %w", err)
	}

	result := fmt.Sprintf("Successfully copied %d tabs from Android device (format: %s):\n\n%s", len(tabs), outputFormat, formattedTabs)
	return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
}

// copyTabsIOS implements the iOS tab copying tool
func (s *TabTransferServer) copyTabsIOS(args IOSTabsArgs) (*mcp_golang.ToolResponse, error) {
	// Set defaults
	if args.Port == 0 {
		args.Port = 9222
	}
	if args.Timeout == 0 {
		args.Timeout = 10
	}
	if args.Wait == 0 {
		args.Wait = 2
	}

	config := driver.IOSConfig{
		DriverConfig: driver.DriverConfig{
			Port:    args.Port,
			Timeout: time.Duration(args.Timeout) * time.Second,
			Debug:   args.Debug,
		},
		Wait: time.Duration(args.Wait) * time.Second,
	}

	iosDriver := driver.NewIOSDriver(config)
	
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(args.Timeout+10)*time.Second)
	defer cancel()

	// Start driver
	if err := iosDriver.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start iOS driver: %w", err)
	}
	defer iosDriver.Stop(ctx)

	// Load tabs
	tabs, err := iosDriver.LoadTabs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load tabs: %w", err)
	}

	// Determine output format
	outputFormat := format.FormatJSON
	if args.Format != "" {
		if parsedFormat, err := format.ParseFormat(args.Format); err == nil {
			outputFormat = parsedFormat
		}
	}

	// Format tabs according to specified format
	formatter := format.NewTabFormatter(outputFormat)
	formattedTabs, err := formatter.FormatTabs(tabs)
	if err != nil {
		return nil, fmt.Errorf("failed to format tabs: %w", err)
	}

	result := fmt.Sprintf("Successfully copied %d tabs from iOS device (format: %s):\n\n%s", len(tabs), outputFormat, formattedTabs)
	return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
}

// reopenTabs implements the tab restoration tool
func (s *TabTransferServer) reopenTabs(args ReopenTabsArgs) (*mcp_golang.ToolResponse, error) {
	// Parse tabs JSON
	var tabs []loader.Tab
	if err := json.Unmarshal([]byte(args.TabsJSON), &tabs); err != nil {
		return nil, fmt.Errorf("failed to parse tabs JSON: %w", err)
	}

	// Set defaults
	if args.Port == 0 {
		args.Port = 9222
	}
	if args.Timeout == 0 {
		args.Timeout = 10
	}

	timeout := time.Duration(args.Timeout) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout+10*time.Second)
	defer cancel()

	var err error
	var result string

	switch args.Platform {
	case "android":
		config := driver.AndroidConfig{
			DriverConfig: driver.DriverConfig{
				Port:    args.Port,
				Timeout: timeout,
				Debug:   args.Debug,
			},
			Socket: "chrome_devtools_remote",
			Wait:   2 * time.Second,
		}
		
		androidDriver := driver.NewAndroidDriver(config)
		if err = androidDriver.Start(ctx); err != nil {
			return nil, fmt.Errorf("failed to start Android driver: %w", err)
		}
		defer androidDriver.Stop(ctx)
		
		if err = androidDriver.RestoreTabs(ctx, tabs); err != nil {
			return nil, fmt.Errorf("failed to restore tabs: %w", err)
		}
		
		result = fmt.Sprintf("Successfully restored %d tabs to Android device", len(tabs))

	case "ios":
		config := driver.IOSConfig{
			DriverConfig: driver.DriverConfig{
				Port:    args.Port,
				Timeout: timeout,
				Debug:   args.Debug,
			},
			Wait: 2 * time.Second,
		}
		
		iosDriver := driver.NewIOSDriver(config)
		if err = iosDriver.Start(ctx); err != nil {
			return nil, fmt.Errorf("failed to start iOS driver: %w", err)
		}
		defer iosDriver.Stop(ctx)
		
		if err = iosDriver.RestoreTabs(ctx, tabs); err != nil {
			return nil, fmt.Errorf("failed to restore tabs: %w", err)
		}
		
		result = fmt.Sprintf("Successfully initiated restoration of %d tabs to iOS device via WebSocket client", len(tabs))

	default:
		return nil, fmt.Errorf("unsupported platform: %s (use 'android' or 'ios')", args.Platform)
	}

	return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
}

// checkEnvironment implements the environment checking tool
func (s *TabTransferServer) checkEnvironment(args CheckEnvironmentArgs) (*mcp_golang.ToolResponse, error) {
	results := make(map[string]string)

	checkPlatform := args.Platform
	if checkPlatform == "" {
		checkPlatform = "all"
	}

	if checkPlatform == "all" || checkPlatform == "android" {
		// Check ADB installation
		if err := platform.CheckADBAvailable(); err != nil {
			results["android_adb"] = fmt.Sprintf("❌ ADB: %v", err)
		} else {
			results["android_adb"] = "✅ ADB: Available and working"
		}
		
		// Check Android device connection
		if err := platform.CheckADBDeviceConnected(); err != nil {
			results["android_device"] = fmt.Sprintf("❌ Android Device: %v", err)
		} else {
			results["android_device"] = "✅ Android Device: Connected and authorized"
		}
	}

	if checkPlatform == "all" || checkPlatform == "ios" {
		// Check iOS WebKit Debug Proxy installation
		if err := platform.CheckIOSWebKitDebugProxyAvailable(); err != nil {
			results["ios_proxy"] = fmt.Sprintf("❌ iOS WebKit Debug Proxy: %v", err)
		} else {
			results["ios_proxy"] = "✅ iOS WebKit Debug Proxy: Available and working"
		}
		
		// iOS device connection check (informational)
		if err := platform.CheckIOSDeviceConnected(); err != nil {
			results["ios_device"] = fmt.Sprintf("ℹ️ iOS Device: %v", err)
		} else {
			results["ios_device"] = "✅ iOS Device: Connection verified"
		}
	}

	// Format results
	resultText := "Environment Check Results:\n\n"
	for _, status := range results {
		resultText += fmt.Sprintf("%s\n", status)
	}
	
	// Add quick fix suggestions
	hasErrors := strings.Contains(resultText, "❌")
	if hasErrors {
		resultText += "\n🔧 Quick Fixes:\n"
		if strings.Contains(resultText, "adb command not found") {
			resultText += "• Install ADB: Run the installation command shown above\n"
		}
		if strings.Contains(resultText, "no Android devices found") {
			resultText += "• Connect Android device and enable USB debugging\n"
		}
		if strings.Contains(resultText, "unauthorized") {
			resultText += "• Check Android device screen for USB debugging prompt\n"
		}
		if strings.Contains(resultText, "ios_webkit_debug_proxy command not found") {
			resultText += "• Install iOS WebKit Debug Proxy: Run the installation command shown above\n"
		}
	} else {
		resultText += "\n✅ All systems ready for tab transfer operations!"
	}

	return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(resultText)), nil
}

// RefreshTabCacheArgs represents arguments for cache refresh
type RefreshTabCacheArgs struct {
	// No arguments needed for cache refresh
}

// refreshTabCache implements the tab cache refresh tool
func (s *TabTransferServer) refreshTabCache(args RefreshTabCacheArgs) (*mcp_golang.ToolResponse, error) {
	if err := s.fetchAndCacheAndroidTabs(); err != nil {
		return nil, fmt.Errorf("failed to refresh tab cache: %w", err)
	}
	
	s.cacheMutex.RLock()
	cacheCount := len(s.tabCache)
	lastUpdate := s.lastUpdated.Format("2006-01-02 15:04:05")
	s.cacheMutex.RUnlock()
	
	result := fmt.Sprintf("✅ Tab cache refreshed successfully!\n\nCached %d tabs\nLast updated: %s", cacheCount, lastUpdate)
	return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
}

// CacheStatusArgs represents arguments for cache status checking
type CacheStatusArgs struct {
	// No arguments needed for cache status
}

// CloseTabArgs represents arguments for single tab closing
type CloseTabArgs struct {
	TabId    string `json:"tabId" jsonschema:"required,description=Unique tab ID to close"`
	Platform string `json:"platform" jsonschema:"description=Target platform: android or ios (default: android)"`
	Confirm  bool   `json:"confirm" jsonschema:"description=Skip confirmation prompt (default: false)"`
}

// CloseTabsBulkArgs represents arguments for bulk tab closing
type CloseTabsBulkArgs struct {
	TabIds      []string `json:"tabIds" jsonschema:"description=Array of specific tab IDs to close"`
	Platform    string   `json:"platform" jsonschema:"description=Target platform: android or ios (default: android)"`
	FilterUrl   string   `json:"filterUrl" jsonschema:"description=Close tabs matching URL pattern (supports wildcards)"`
	FilterTitle string   `json:"filterTitle" jsonschema:"description=Close tabs matching title pattern (supports wildcards)"`
	Confirm     bool     `json:"confirm" jsonschema:"description=Skip confirmation prompt (default: false)"`
	DryRun      bool     `json:"dryRun" jsonschema:"description=Preview operation without actually closing tabs (default: false)"`
}

// SearchTabsArgs represents arguments for tab searching
type SearchTabsArgs struct {
	Query  string `json:"query" jsonschema:"description=Search query to match against URLs and titles"`
	Domain string `json:"domain" jsonschema:"description=Filter by specific domain (e.g. github.com)"`
	Title  string `json:"title" jsonschema:"description=Search specifically in tab titles"`
	URL    string `json:"url" jsonschema:"description=Search specifically in URLs"`
	Limit  int    `json:"limit" jsonschema:"description=Maximum number of results to return (default: 10)"`
	Format string `json:"format" jsonschema:"description=Output format: json or yaml (default: json)"`
}

// cacheStatus implements the cache status tool
func (s *TabTransferServer) cacheStatus(args CacheStatusArgs) (*mcp_golang.ToolResponse, error) {
	s.cacheMutex.RLock()
	cacheCount := len(s.tabCache)
	cacheSize := s.cacheSize
	lastUpdate := s.lastUpdated
	s.cacheMutex.RUnlock()
	
	var statusText strings.Builder
	statusText.WriteString("📊 Tab Cache Status\n\n")
	statusText.WriteString(fmt.Sprintf("📱 Cached Tabs: %d\n", cacheCount))
	statusText.WriteString(fmt.Sprintf("🎯 Cache Limit: %d\n", cacheSize))
	
	if lastUpdate.IsZero() {
		statusText.WriteString("⏰ Last Updated: Never (cache not populated)\n")
		statusText.WriteString("📊 Status: Empty - use refresh_tab_cache tool to populate\n")
	} else {
		statusText.WriteString(fmt.Sprintf("⏰ Last Updated: %s\n", lastUpdate.Format("2006-01-02 15:04:05")))
		statusText.WriteString(fmt.Sprintf("📊 Status: Active (%d/%d tabs)\n", cacheCount, cacheSize))
		
		// Show age of cache
		age := time.Since(lastUpdate)
		if age < time.Minute {
			statusText.WriteString("🟢 Cache Age: Fresh (< 1 minute)\n")
		} else if age < time.Hour {
			statusText.WriteString(fmt.Sprintf("🟡 Cache Age: %d minutes\n", int(age.Minutes())))
		} else {
			statusText.WriteString(fmt.Sprintf("🔴 Cache Age: %d hours\n", int(age.Hours())))
		}
	}
	
	return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(statusText.String())), nil
}

// getCurrentTabsYAML implements the current tabs resource (YAML format only)
func (s *TabTransferServer) getCurrentTabsYAML() (*mcp_golang.ResourceResponse, error) {
	// Return cached tabs with thread safety
	s.cacheMutex.RLock()
	cachedTabs := make([]loader.Tab, len(s.tabCache))
	copy(cachedTabs, s.tabCache)
	s.cacheMutex.RUnlock()
	
	formatter := format.YAMLFormatter()
	tabsData, err := formatter.FormatTabs(cachedTabs)
	if err != nil {
		return nil, fmt.Errorf("failed to format cached tabs as YAML: %w", err)
	}
	
	resource := mcp_golang.NewTextEmbeddedResource(
		"tabs://current",
		tabsData,
		formatter.GetMimeType(),
	)
	
	return mcp_golang.NewResourceResponse(resource), nil
}

// closeTab implements the single tab closing tool
func (s *TabTransferServer) closeTab(args CloseTabArgs) (*mcp_golang.ToolResponse, error) {
	// Default platform to android
	platform := args.Platform
	if platform == "" {
		platform = "android"
	}
	
	// Validation
	if args.TabId == "" {
		return nil, fmt.Errorf("tabId is required")
	}
	
	// Safety confirmation (unless explicitly confirmed)
	if !args.Confirm {
		confirmText := fmt.Sprintf("⚠️ WARNING: You are about to permanently close tab:\nID: %s\nPlatform: %s\n\nThis action cannot be undone. To proceed, call this tool again with confirm=true.", args.TabId, platform)
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(confirmText)), nil
	}
	
	// Only Android is supported for now
	if platform != "android" {
		return nil, fmt.Errorf("tab closing is currently only supported for Android platform")
	}
	
	// Setup Android driver
	config := driver.AndroidConfig{
		DriverConfig: driver.DriverConfig{
			Port:    9222,
			Timeout: 10 * time.Second,
			Debug:   true,
		},
		Socket: "chrome_devtools_remote",
		Wait:   2 * time.Second,
	}
	
	androidDriver := driver.NewAndroidDriver(config)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	
	// Start driver
	if err := androidDriver.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start Android driver: %w", err)
	}
	defer androidDriver.Stop(ctx)
	
	// Close the tab
	if err := androidDriver.CloseTab(ctx, args.TabId); err != nil {
		return nil, fmt.Errorf("failed to close tab: %w", err)
	}
	
	result := fmt.Sprintf("✅ Successfully closed tab: %s", args.TabId)
	return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
}

// closeTabsBulk implements the bulk tab closing tool
func (s *TabTransferServer) closeTabsBulk(args CloseTabsBulkArgs) (*mcp_golang.ToolResponse, error) {
	// Default platform to android
	platform := args.Platform
	if platform == "" {
		platform = "android"
	}
	
	// Only Android is supported for now
	if platform != "android" {
		return nil, fmt.Errorf("bulk tab closing is currently only supported for Android platform")
	}
	
	// Setup Android driver
	config := driver.AndroidConfig{
		DriverConfig: driver.DriverConfig{
			Port:    9222,
			Timeout: 10 * time.Second,
			Debug:   args.DryRun, // Enable debug for dry run to see what would happen
		},
		Socket: "chrome_devtools_remote",
		Wait:   2 * time.Second,
	}
	
	androidDriver := driver.NewAndroidDriver(config)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	// Start driver
	if err := androidDriver.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start Android driver: %w", err)
	}
	defer androidDriver.Stop(ctx)
	
	// Load current tabs to apply filters
	currentTabs, err := androidDriver.LoadTabs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load current tabs: %w", err)
	}
	
	// Determine which tabs to close
	var tabsToClose []string
	
	if len(args.TabIds) > 0 {
		// Use provided tab IDs
		tabsToClose = args.TabIds
	} else {
		// Apply filters to find tabs to close
		for _, tab := range currentTabs {
			shouldClose := true
			
			// Apply URL filter if provided
			if args.FilterUrl != "" {
				if !matchesPattern(tab.URL, args.FilterUrl) {
					shouldClose = false
				}
			}
			
			// Apply title filter if provided
			if args.FilterTitle != "" {
				if !matchesPattern(tab.Title, args.FilterTitle) {
					shouldClose = false
				}
			}
			
			if shouldClose {
				tabsToClose = append(tabsToClose, tab.ID)
			}
		}
	}
	
	if len(tabsToClose) == 0 {
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent("No tabs match the specified criteria.")), nil
	}
	
	// Dry run: just show what would be closed
	if args.DryRun {
		var preview strings.Builder
		preview.WriteString(fmt.Sprintf("🔍 DRY RUN: Would close %d tabs:\n\n", len(tabsToClose)))
		
		for _, tabID := range tabsToClose {
			// Find the tab details
			for _, tab := range currentTabs {
				if tab.ID == tabID {
					preview.WriteString(fmt.Sprintf("• %s\n  ID: %s\n  URL: %s\n\n", tab.Title, tab.ID, tab.URL))
					break
				}
			}
		}
		
		preview.WriteString("To actually close these tabs, call this tool again with dryRun=false and confirm=true.")
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(preview.String())), nil
	}
	
	// Safety confirmation (unless explicitly confirmed)
	if !args.Confirm {
		confirmText := fmt.Sprintf("⚠️ WARNING: You are about to permanently close %d tabs on %s.\n\nThis action cannot be undone. To proceed, call this tool again with confirm=true.\n\nTip: Use dryRun=true first to preview which tabs will be closed.", len(tabsToClose), platform)
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(confirmText)), nil
	}
	
	// Actually close the tabs
	if err := androidDriver.CloseTabs(ctx, tabsToClose); err != nil {
		return nil, fmt.Errorf("failed to close tabs: %w", err)
	}
	
	result := fmt.Sprintf("✅ Successfully closed %d tabs", len(tabsToClose))
	return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(result)), nil
}

// matchesPattern checks if a string matches a pattern (supports wildcards)
func matchesPattern(text, pattern string) bool {
	// Simple wildcard matching - supports * as wildcard
	if pattern == "*" {
		return true
	}
	
	// For now, simple contains check - can be enhanced later
	return strings.Contains(strings.ToLower(text), strings.ToLower(pattern))
}


// searchTabs implements the tab search tool
func (s *TabTransferServer) searchTabs(args SearchTabsArgs) (*mcp_golang.ToolResponse, error) {
	// Set defaults
	if args.Limit == 0 {
		args.Limit = 10
	}
	
	// Get cached tabs
	s.cacheMutex.RLock()
	cachedTabs := make([]loader.Tab, len(s.tabCache))
	copy(cachedTabs, s.tabCache)
	s.cacheMutex.RUnlock()
	
	if len(cachedTabs) == 0 {
		return mcp_golang.NewToolResponse(mcp_golang.NewTextContent("No tabs are currently cached. Use refresh_tab_cache tool to populate cache first.")), nil
	}
	
	// Apply filters and calculate relevance scores
	var results []format.SearchResult
	
	for _, tab := range cachedTabs {
		score := 0.0
		matches := true
		
		// Apply domain filter
		if args.Domain != "" {
			if !strings.Contains(strings.ToLower(tab.URL), strings.ToLower(args.Domain)) {
				matches = false
			} else {
				score += 2.0 // Domain match gets high score
			}
		}
		
		// Apply title filter
		if args.Title != "" {
			if !strings.Contains(strings.ToLower(tab.Title), strings.ToLower(args.Title)) {
				matches = false
			} else {
				score += 1.5 // Title match gets medium-high score
			}
		}
		
		// Apply URL filter
		if args.URL != "" {
			if !strings.Contains(strings.ToLower(tab.URL), strings.ToLower(args.URL)) {
				matches = false
			} else {
				score += 1.0 // URL match gets medium score
			}
		}
		
		// Apply general query filter
		if args.Query != "" {
			queryLower := strings.ToLower(args.Query)
			titleMatch := strings.Contains(strings.ToLower(tab.Title), queryLower)
			urlMatch := strings.Contains(strings.ToLower(tab.URL), queryLower)
			
			if !titleMatch && !urlMatch {
				matches = false
			} else {
				if titleMatch {
					score += 1.0
				}
				if urlMatch {
					score += 0.5
				}
				
				// Bonus for exact matches
				if strings.EqualFold(tab.Title, args.Query) {
					score += 2.0
				}
				
				// Bonus for query appearing at the start
				if strings.HasPrefix(strings.ToLower(tab.Title), queryLower) {
					score += 1.0
				}
			}
		}
		
		// If no filters provided, include all tabs with minimal score
		if args.Query == "" && args.Domain == "" && args.Title == "" && args.URL == "" {
			matches = true
			score = 0.1
		}
		
		if matches {
			results = append(results, format.SearchResult{
				Tab:   tab,
				Score: score,
			})
		}
	}
	
	// Sort by relevance score (descending)
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[i].Score < results[j].Score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
	
	// Limit results
	if len(results) > args.Limit {
		results = results[:args.Limit]
	}
	
	// Determine output format
	outputFormat := format.FormatJSON
	if args.Format != "" {
		if parsedFormat, err := format.ParseFormat(args.Format); err == nil {
			outputFormat = parsedFormat
		}
	}
	
	// Format output
	var resultText string
	if outputFormat == format.FormatYAML {
		yamlData, err := format.YAMLFormatter().FormatSearchResults(results)
		if err != nil {
			return nil, fmt.Errorf("failed to format search results as YAML: %w", err)
		}
		resultText = fmt.Sprintf("🔍 Found %d tabs matching search criteria (format: yaml):\n\n%s", len(results), yamlData)
	} else {
		jsonData, err := format.JSONFormatter().FormatSearchResults(results)
		if err != nil {
			return nil, fmt.Errorf("failed to format search results as JSON: %w", err)
		}
		resultText = fmt.Sprintf("🔍 Found %d tabs matching search criteria (format: json):\n\n%s", len(results), jsonData)
	}
	
	return mcp_golang.NewToolResponse(mcp_golang.NewTextContent(resultText)), nil
}