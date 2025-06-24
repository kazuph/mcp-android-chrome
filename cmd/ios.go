package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/kazuph/mcp-android-chrome/internal/driver"
)

var iosCmd = &cobra.Command{
	Use:   "ios",
	Short: "Copy tabs from iOS Chrome/Safari via WebKit Debug Proxy",
	Long: `Copy all open tabs from Chrome or Safari on iOS to your computer using iOS WebKit Debug Proxy.

Requirements:
- iOS device with Web Inspector enabled in Safari settings
- Chrome with Web Inspector enabled (iOS 16.4+)
- ios_webkit_debug_proxy installed and in PATH
- USB connection between device and computer

This command will:
1. Start ios_webkit_debug_proxy as background process
2. Connect to WebKit Debug Protocol on device
3. Retrieve all open tabs
4. Output tab information as JSON`,
	Run: func(cmd *cobra.Command, args []string) {
		port, _ := cmd.Flags().GetInt("port")
		timeout, _ := cmd.Flags().GetInt("timeout")
		wait, _ := cmd.Flags().GetInt("wait")
		debug, _ := cmd.Flags().GetBool("debug")

		config := driver.IOSConfig{
			DriverConfig: driver.DriverConfig{
				Port:    port,
				Timeout: time.Duration(timeout) * time.Second,
				Debug:   debug,
			},
			Wait: time.Duration(wait) * time.Second,
		}

		iosDriver := driver.NewIOSDriver(config)
		
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout+10)*time.Second)
		defer cancel()

		fmt.Println("Starting iOS Chrome/Safari tab copy...")

		// Start driver
		if err := iosDriver.Start(ctx); err != nil {
			fmt.Printf("Error: Failed to start iOS driver: %v\n", err)
			return
		}
		defer iosDriver.Stop(ctx)

		// Load tabs
		tabs, err := iosDriver.LoadTabs(ctx)
		if err != nil {
			fmt.Printf("Error: Failed to load tabs: %v\n", err)
			return
		}

		// Output results
		fmt.Printf("Successfully copied %d tabs from iOS device:\n\n", len(tabs))
		
		tabsJSON, err := json.MarshalIndent(tabs, "", "  ")
		if err != nil {
			fmt.Printf("Error: Failed to format tabs: %v\n", err)
			return
		}
		
		fmt.Println(string(tabsJSON))
	},
}

func init() {
	iosCmd.Flags().IntP("port", "p", 9222, "Port for iOS WebKit Debug Proxy")
	iosCmd.Flags().IntP("timeout", "t", 10, "Network timeout in seconds")
	iosCmd.Flags().IntP("wait", "w", 2, "Wait time before starting in seconds")
	iosCmd.Flags().Bool("debug", false, "Enable debug output")
}