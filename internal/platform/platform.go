package platform

import (
	"fmt"
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

// CheckADBAvailable checks if ADB is available and working
func CheckADBAvailable() error {
	if !IsShellCommandAvailable("adb") {
		return fmt.Errorf("adb command not found in PATH. Install with:\n- macOS: brew install --cask android-platform-tools\n- Linux: sudo apt install android-tools-adb\n- Windows: Download from developer.android.com/tools/releases/platform-tools")
	}
	
	// Test adb version to ensure it's working
	cmd := exec.Command("adb", "version")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("adb command failed: %v. Try restarting ADB with 'adb kill-server && adb start-server'", err)
	}
	
	if !strings.Contains(string(output), "Android Debug Bridge") {
		return fmt.Errorf("adb command did not return expected version output")
	}
	
	return nil
}

// CheckADBDeviceConnected checks if any Android devices are connected
func CheckADBDeviceConnected() error {
	cmd := exec.Command("adb", "devices")
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
	cmd := exec.Command("ios_webkit_debug_proxy", "--help")
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