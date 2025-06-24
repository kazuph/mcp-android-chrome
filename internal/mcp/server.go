package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	mcp_golang "github.com/metoro-io/mcp-golang"
	"github.com/metoro-io/mcp-golang/transport/stdio"

	"github.com/kazuph/mcp-android-chrome/internal/driver"
	"github.com/kazuph/mcp-android-chrome/internal/loader"
	"github.com/kazuph/mcp-android-chrome/internal/platform"
)

// TabTransferServer implements MCP server for tab transfer functionality
type TabTransferServer struct {
	server *mcp_golang.Server
}

// NewTabTransferServer creates a new MCP server for tab transfer
func NewTabTransferServer() *TabTransferServer {
	server := mcp_golang.NewServer(stdio.NewStdioServerTransport())
	return &TabTransferServer{
		server: server,
	}
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

	// Start the server
	return s.server.Serve()
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

	return nil
}

// registerResources registers MCP resources
func (s *TabTransferServer) registerResources() error {
	// Resource 1: Current tabs (if any cached)
	err := s.server.RegisterResource("tabs://current", "current_tabs", "Currently loaded tabs", "application/json", s.getCurrentTabs)
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
}

// IOSTabsArgs represents arguments for iOS tab copying
type IOSTabsArgs struct {
	Port    int  `json:"port" jsonschema:"description=Port for iOS WebKit Debug Proxy (default: 9222)"`
	Timeout int  `json:"timeout" jsonschema:"description=Network timeout in seconds (default: 10)"`
	Wait    int  `json:"wait" jsonschema:"description=Wait time before starting in seconds (default: 2)"`
	Debug   bool `json:"debug" jsonschema:"description=Enable debug output"`
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
	Platform string `json:"platform" jsonschema:"description=Specific platform to check (android, ios, or all)"`
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

	// Convert to JSON for response
	tabsJSON, err := json.MarshalIndent(tabs, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tabs: %w", err)
	}

	result := fmt.Sprintf("Successfully copied %d tabs from Android device:\n\n%s", len(tabs), string(tabsJSON))
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

	// Convert to JSON for response
	tabsJSON, err := json.MarshalIndent(tabs, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tabs: %w", err)
	}

	result := fmt.Sprintf("Successfully copied %d tabs from iOS device:\n\n%s", len(tabs), string(tabsJSON))
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

// getCurrentTabs implements the current tabs resource
func (s *TabTransferServer) getCurrentTabs() (*mcp_golang.ResourceResponse, error) {
	// This would return cached tabs if any exist
	// For now, return an empty response
	emptyTabs := []loader.Tab{}
	tabsJSON, _ := json.MarshalIndent(emptyTabs, "", "  ")
	
	resource := mcp_golang.NewTextEmbeddedResource(
		"tabs://current",
		string(tabsJSON),
		"application/json",
	)
	
	return mcp_golang.NewResourceResponse(resource), nil
}