package cmd

import (
	"fmt"
	"os"

	"github.com/kazuph/mcp-android-chrome/internal/mcp"
	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start MCP server",
	Long: `Start the Model Context Protocol server that provides tab transfer
functionality to AI assistants like Claude.

The server exposes tools for:
- copy_tabs_android: Copy tabs from Android Chrome
- copy_tabs_ios: Copy tabs from iOS Chrome/Safari  
- reopen_tabs: Restore saved tabs to mobile devices
- check_environment: Verify system dependencies

Configure in Claude Desktop's claude_desktop_config.json:
{
  "mcpServers": {
    "android-chrome": {
      "command": "/path/to/mcp-android-chrome",
      "args": ["mcp"]
    }
  }
}`,
	Run: func(cmd *cobra.Command, args []string) {
		// Don't print anything to stdout - MCP uses stdio for JSON-RPC communication
		// Any debug output should go to stderr instead
		
		server := mcp.NewTabTransferServer()
		if err := server.Start(); err != nil {
			// Use stderr for error messages in MCP mode
			fmt.Fprintf(os.Stderr, "Failed to start MCP server: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	// MCP server flags can be added here if needed
}