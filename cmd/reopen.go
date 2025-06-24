package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/kazuph/mcp-android-chrome/internal/driver"
	"github.com/kazuph/mcp-android-chrome/internal/loader"
)

var reopenCmd = &cobra.Command{
	Use:   "reopen [tabs-file.json]",
	Short: "Restore saved tabs to mobile device",
	Long: `Restore previously saved tabs to an Android or iOS mobile device.

This command reads a JSON file containing tab information and restores
those tabs to the specified mobile device platform.

For Android:
- Uses ADB and Chrome DevTools Protocol
- Creates tabs via HTTP API

For iOS:
- Uses iOS WebKit Debug Proxy and WebSocket connection
- Creates an HTML interface for tab restoration

Examples:
  mcp-android-chrome reopen --platform android tabs.json
  mcp-android-chrome reopen --platform ios --port 9222 saved-tabs.json`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		platform, _ := cmd.Flags().GetString("platform")
		port, _ := cmd.Flags().GetInt("port")
		timeout, _ := cmd.Flags().GetInt("timeout")
		debug, _ := cmd.Flags().GetBool("debug")

		if platform == "" {
			fmt.Println("Error: --platform flag is required (android or ios)")
			return
		}

		// Read tabs from file
		tabsFile := args[0]
		tabsData, err := os.ReadFile(tabsFile)
		if err != nil {
			fmt.Printf("Error: Failed to read tabs file: %v\n", err)
			return
		}

		var tabs []loader.Tab
		if err := json.Unmarshal(tabsData, &tabs); err != nil {
			fmt.Printf("Error: Failed to parse tabs JSON: %v\n", err)
			return
		}

		fmt.Printf("Restoring %d tabs to %s device...\n", len(tabs), platform)

		timeout_duration := time.Duration(timeout) * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), timeout_duration+10*time.Second)
		defer cancel()

		switch platform {
		case "android":
			if err := restoreAndroidTabs(ctx, tabs, port, timeout_duration, debug); err != nil {
				fmt.Printf("Error: Failed to restore Android tabs: %v\n", err)
				return
			}
			fmt.Printf("Successfully restored %d tabs to Android device\n", len(tabs))

		case "ios":
			if err := restoreIOSTabs(ctx, tabs, port, timeout_duration, debug); err != nil {
				fmt.Printf("Error: Failed to restore iOS tabs: %v\n", err)
				return
			}
			fmt.Printf("Successfully initiated restoration of %d tabs to iOS device\n", len(tabs))

		default:
			fmt.Printf("Error: Unsupported platform: %s (use 'android' or 'ios')\n", platform)
			return
		}
	},
}

func restoreAndroidTabs(ctx context.Context, tabs []loader.Tab, port int, timeout time.Duration, debug bool) error {
	config := driver.AndroidConfig{
		DriverConfig: driver.DriverConfig{
			Port:    port,
			Timeout: timeout,
			Debug:   debug,
		},
		Socket: "chrome_devtools_remote",
		Wait:   2 * time.Second,
	}
	
	androidDriver := driver.NewAndroidDriver(config)
	if err := androidDriver.Start(ctx); err != nil {
		return fmt.Errorf("failed to start Android driver: %w", err)
	}
	defer androidDriver.Stop(ctx)
	
	return androidDriver.RestoreTabs(ctx, tabs)
}

func restoreIOSTabs(ctx context.Context, tabs []loader.Tab, port int, timeout time.Duration, debug bool) error {
	config := driver.IOSConfig{
		DriverConfig: driver.DriverConfig{
			Port:    port,
			Timeout: timeout,
			Debug:   debug,
		},
		Wait: 2 * time.Second,
	}
	
	iosDriver := driver.NewIOSDriver(config)
	if err := iosDriver.Start(ctx); err != nil {
		return fmt.Errorf("failed to start iOS driver: %w", err)
	}
	defer iosDriver.Stop(ctx)
	
	return iosDriver.RestoreTabs(ctx, tabs)
}

func init() {
	reopenCmd.Flags().StringP("platform", "P", "", "Target platform (android or ios) [required]")
	reopenCmd.Flags().IntP("port", "p", 9222, "Port for device communication")
	reopenCmd.Flags().IntP("timeout", "t", 10, "Network timeout in seconds")
	reopenCmd.Flags().Bool("debug", false, "Enable debug output")
	reopenCmd.MarkFlagRequired("platform")
}