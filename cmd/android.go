package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/kazuph/mcp-android-chrome/internal/driver"
)

var androidCmd = &cobra.Command{
	Use:   "android",
	Short: "Copy tabs from Android Chrome via ADB",
	Long: `Copy all open tabs from Chrome on Android to your computer using ADB.

Requirements:
- Android device with USB debugging enabled
- ADB (Android Debug Bridge) installed and in PATH
- Chrome browser running on Android device
- USB connection between device and computer

This command will:
1. Setup ADB port forwarding
2. Connect to Chrome DevTools Protocol on device
3. Retrieve all open tabs
4. Output tab information as JSON`,
	Run: func(cmd *cobra.Command, args []string) {
		port, _ := cmd.Flags().GetInt("port")
		socket, _ := cmd.Flags().GetString("socket")
		timeout, _ := cmd.Flags().GetInt("timeout")
		wait, _ := cmd.Flags().GetInt("wait")
		skipCleanup, _ := cmd.Flags().GetBool("skip-cleanup")
		debug, _ := cmd.Flags().GetBool("debug")

		config := driver.AndroidConfig{
			DriverConfig: driver.DriverConfig{
				Port:    port,
				Timeout: time.Duration(timeout) * time.Second,
				Debug:   debug,
			},
			Socket:      socket,
			Wait:        time.Duration(wait) * time.Second,
			SkipCleanup: skipCleanup,
		}

		androidDriver := driver.NewAndroidDriver(config)
		
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout+10)*time.Second)
		defer cancel()

		fmt.Println("Starting Android Chrome tab copy...")

		// Start driver
		if err := androidDriver.Start(ctx); err != nil {
			fmt.Printf("Error: Failed to start Android driver: %v\n", err)
			return
		}
		defer androidDriver.Stop(ctx)

		// Load tabs
		tabs, err := androidDriver.LoadTabs(ctx)
		if err != nil {
			fmt.Printf("Error: Failed to load tabs: %v\n", err)
			return
		}

		// Output results
		fmt.Printf("Successfully copied %d tabs from Android device:\n\n", len(tabs))
		
		tabsJSON, err := json.MarshalIndent(tabs, "", "  ")
		if err != nil {
			fmt.Printf("Error: Failed to format tabs: %v\n", err)
			return
		}
		
		fmt.Println(string(tabsJSON))
	},
}

func init() {
	androidCmd.Flags().IntP("port", "p", 9222, "Port for ADB forwarding")
	androidCmd.Flags().StringP("socket", "s", "chrome_devtools_remote", "ADB socket name")
	androidCmd.Flags().IntP("timeout", "t", 10, "Network timeout in seconds")
	androidCmd.Flags().IntP("wait", "w", 2, "Wait time before starting in seconds")
	androidCmd.Flags().Bool("skip-cleanup", false, "Skip ADB cleanup after operation")
	androidCmd.Flags().Bool("debug", false, "Enable debug output")
}