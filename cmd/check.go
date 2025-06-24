package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	platformpkg "github.com/kazuph/mcp-android-chrome/internal/platform"
)

var checkCmd = &cobra.Command{
	Use:   "check [platform]",
	Short: "Check system dependencies",
	Long: `Check that required system dependencies are installed and available.

This command verifies:
- ADB (Android Debug Bridge) for Android support
- iOS WebKit Debug Proxy for iOS support

You can check a specific platform or all platforms:
  mcp-android-chrome check           # Check all platforms
  mcp-android-chrome check android   # Check only Android dependencies
  mcp-android-chrome check ios       # Check only iOS dependencies`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		platform := "all"
		if len(args) > 0 {
			platform = args[0]
		}

		fmt.Println("Checking system dependencies...")
		fmt.Printf("Platform: %s\n\n", platform)

		hasErrors := false

		if platform == "all" || platform == "android" {
			fmt.Print("Android (ADB): ")
			if err := platformpkg.CheckADBAvailable(); err != nil {
				fmt.Printf("❌ %v\n", err)
				hasErrors = true
			} else {
				fmt.Println("✅ Available and working")
			}
		}

		if platform == "all" || platform == "ios" {
			fmt.Print("iOS (WebKit Debug Proxy): ")
			if err := platformpkg.CheckIOSWebKitDebugProxyAvailable(); err != nil {
				fmt.Printf("❌ %v\n", err)
				hasErrors = true
			} else {
				fmt.Println("✅ Available and working")
			}
		}

		if platform != "all" && platform != "android" && platform != "ios" {
			fmt.Printf("Error: Unknown platform '%s'. Use 'android', 'ios', or omit for all.\n", platform)
			return
		}

		fmt.Println()
		if hasErrors {
			fmt.Println("❌ Some dependencies are missing. Please install them before using this tool.")
			fmt.Println()
			fmt.Println("Installation instructions:")
			if platform == "all" || platform == "android" {
				fmt.Println("Android:")
				fmt.Println("  macOS: brew install --cask android-platform-tools")
				fmt.Println("  Linux: sudo apt install android-tools-adb")
				fmt.Println("  Windows: Download from https://developer.android.com/tools/releases/platform-tools")
			}
			if platform == "all" || platform == "ios" {
				fmt.Println("iOS:")
				fmt.Println("  macOS: brew install ios-webkit-debug-proxy")
				fmt.Println("  Linux: See https://github.com/google/ios-webkit-debug-proxy")
				fmt.Println("  Windows: Not officially supported")
			}
		} else {
			fmt.Println("✅ All required dependencies are available!")
		}
	},
}