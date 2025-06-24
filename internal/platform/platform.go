package platform

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// IsWindows returns true if running on Windows
func IsWindows() bool {
	return runtime.GOOS == "windows"
}

// IsMac returns true if running on macOS
func IsMac() bool {
	return runtime.GOOS == "darwin"
}

// IsLinux returns true if running on Linux
func IsLinux() bool {
	return runtime.GOOS == "linux"
}

// IsShellCommandAvailable checks if a command is available in PATH
func IsShellCommandAvailable(command string) bool {
	cmd := exec.Command("which", command)
	if IsWindows() {
		cmd = exec.Command("where", command)
	}
	
	err := cmd.Run()
	return err == nil
}

// FindADBPath finds the absolute path to ADB command
func FindADBPath() string {
	// First check if ADB_PATH environment variable is set
	if adbPath := os.Getenv("ADB_PATH"); adbPath != "" {
		if _, err := os.Stat(adbPath); err == nil {
			return adbPath
		}
	}
	
	// Try common ADB locations
	commonPaths := []string{
		"/Users/kazuph/Library/Android/sdk/platform-tools/adb",
		"/opt/homebrew/bin/adb",
		"/usr/local/bin/adb",
		"adb", // Fallback to PATH
	}
	
	if IsWindows() {
		commonPaths = []string{
			"C:\\Users\\%USERNAME%\\AppData\\Local\\Android\\Sdk\\platform-tools\\adb.exe",
			"C:\\Android\\platform-tools\\adb.exe",
			"adb.exe",
		}
	}
	
	for _, path := range commonPaths {
		if path == "adb" || path == "adb.exe" {
			// Test if command exists in PATH
			cmd := exec.Command("which", "adb")
			if IsWindows() {
				cmd = exec.Command("where", "adb.exe")
			}
			if cmd.Run() == nil {
				return path
			}
		} else {
			// Test if file exists at absolute path
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	}
	
	return "adb" // Fallback
}

// FindIOSWebKitDebugProxyPath finds the absolute path to ios_webkit_debug_proxy
func FindIOSWebKitDebugProxyPath() string {
	// First check if IOS_WEBKIT_DEBUG_PROXY_PATH environment variable is set
	if proxyPath := os.Getenv("IOS_WEBKIT_DEBUG_PROXY_PATH"); proxyPath != "" {
		if _, err := os.Stat(proxyPath); err == nil {
			return proxyPath
		}
	}
	
	commonPaths := []string{
		"/opt/homebrew/bin/ios_webkit_debug_proxy",
		"/usr/local/bin/ios_webkit_debug_proxy",
		"ios_webkit_debug_proxy", // Fallback to PATH
	}
	
	for _, path := range commonPaths {
		if path == "ios_webkit_debug_proxy" {
			// Test if command exists in PATH
			cmd := exec.Command("which", "ios_webkit_debug_proxy")
			if cmd.Run() == nil {
				return path
			}
		} else {
			// Test if file exists at absolute path
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	}
	
	return "ios_webkit_debug_proxy" // Fallback
}

// CheckADBAvailable checks if ADB is available and working
func CheckADBAvailable() error {
	// Try to find ADB path first
	adbPath := FindADBPath()
	
	// Test adb version to ensure it's working
	cmd := exec.Command(adbPath, "version")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("adb command not found or failed: %v. Install with:\n- macOS: brew install --cask android-platform-tools\n- Linux: sudo apt install android-tools-adb\n- Windows: Download from developer.android.com/tools/releases/platform-tools", err)
	}
	
	if !strings.Contains(string(output), "Android Debug Bridge") {
		return fmt.Errorf("adb command did not return expected version output")
	}
	
	return nil
}

// CheckADBDeviceConnected checks if any Android devices are connected
func CheckADBDeviceConnected() error {
	adbPath := FindADBPath()
	cmd := exec.Command(adbPath, "devices")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to list ADB devices: %v", err)
	}
	
	lines := strings.Split(string(output), "\n")
	deviceCount := 0
	unauthorizedCount := 0
	
	for _, line := range lines[1:] { // Skip "List of devices attached" header
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		if strings.Contains(line, "unauthorized") {
			unauthorizedCount++
		} else if strings.Contains(line, "device") {
			deviceCount++
		}
	}
	
	if deviceCount == 0 && unauthorizedCount == 0 {
		return fmt.Errorf("no Android devices found. Please:\n1. Connect device via USB\n2. Enable USB debugging in Developer Options\n3. Ensure USB cable supports data transfer")
	}
	
	if unauthorizedCount > 0 && deviceCount == 0 {
		return fmt.Errorf("Android device found but unauthorized. Please:\n1. Check device screen for USB debugging prompt\n2. Tap 'Allow' to authorize this computer\n3. Ensure device is unlocked")
	}
	
	return nil
}

// CheckIOSWebKitDebugProxyAvailable checks if ios_webkit_debug_proxy is available
func CheckIOSWebKitDebugProxyAvailable() error {
	if !IsShellCommandAvailable("ios_webkit_debug_proxy") {
		return fmt.Errorf("ios_webkit_debug_proxy command not found in PATH. Install with:\n- macOS: brew install ios-webkit-debug-proxy\n- Linux: See github.com/google/ios-webkit-debug-proxy for build instructions\n- Windows: Not officially supported")
	}
	
	// Test help output to ensure it's working
	proxyPath := FindIOSWebKitDebugProxyPath()
	cmd := exec.Command(proxyPath, "--help")
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("ios_webkit_debug_proxy command failed: %v", err)
	}
	
	return nil
}

// CheckIOSDeviceConnected checks if any iOS devices are connected (basic check)
func CheckIOSDeviceConnected() error {
	// This is a basic check - ios_webkit_debug_proxy doesn't have a simple device list command
	// We can only verify this when actually trying to connect
	return fmt.Errorf("iOS device connectivity can only be verified during actual connection attempt. Ensure:\n1. iOS device connected via USB\n2. Device unlocked and trusted\n3. Web Inspector enabled in Safari settings\n4. For Chrome: Web Inspector enabled in Chrome settings")
}

// OpenInBrowser opens a URL in the default browser
func OpenInBrowser(url string) error {
	var cmd *exec.Cmd
	
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	
	return cmd.Run()
}