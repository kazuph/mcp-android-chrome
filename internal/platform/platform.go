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
		return fmt.Errorf("adb command not found in PATH")
	}
	
	// Test adb version to ensure it's working
	cmd := exec.Command("adb", "version")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("adb command failed: %v", err)
	}
	
	if !strings.Contains(string(output), "Android Debug Bridge") {
		return fmt.Errorf("adb command did not return expected version output")
	}
	
	return nil
}

// CheckIOSWebKitDebugProxyAvailable checks if ios_webkit_debug_proxy is available
func CheckIOSWebKitDebugProxyAvailable() error {
	if !IsShellCommandAvailable("ios_webkit_debug_proxy") {
		return fmt.Errorf("ios_webkit_debug_proxy command not found in PATH")
	}
	
	// Test help output to ensure it's working
	cmd := exec.Command("ios_webkit_debug_proxy", "--help")
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("ios_webkit_debug_proxy command failed: %v", err)
	}
	
	return nil
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