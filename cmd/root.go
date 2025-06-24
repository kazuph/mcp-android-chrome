package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "mcp-android-chrome",
	Short: "A Model Context Protocol server for Android/iOS Chrome tab transfer",
	Long: `mcp-android-chrome is a Go port of the tab-transfer tool that provides
MCP server functionality for transferring Chrome tabs between mobile devices
and computers using developer tools.

This tool supports:
- Copying tabs from Android Chrome via ADB
- Copying tabs from iOS Chrome/Safari via iOS WebKit Debug Proxy
- Reopening saved tabs on mobile devices
- Environment dependency checking

Original tool by machinateur, Go port by kazuph.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(mcpCmd)
	rootCmd.AddCommand(androidCmd)
	rootCmd.AddCommand(iosCmd)
	rootCmd.AddCommand(reopenCmd)
	rootCmd.AddCommand(checkCmd)
}