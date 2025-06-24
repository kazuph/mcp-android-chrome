# MCP Android Chrome

A Model Context Protocol (MCP) server for transferring Chrome tabs between mobile devices and computers using developer tools.

[![Release](https://img.shields.io/github/v/release/kazuph/mcp-android-chrome)](https://github.com/kazuph/mcp-android-chrome/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go](https://img.shields.io/badge/go-%2300ADD8.svg?style=flat&logo=go&logoColor=white)](https://golang.org/)

## Overview

This is a Go port of the original [tab-transfer](https://github.com/machinateur/tab-transfer) tool by **machinateur**, reimplemented as an MCP server to integrate with AI assistants like Claude Desktop.

**Original tool by machinateur, Go port by kazuph.**

## What's New in v0.9.0

- ✅ **Full MCP Server Implementation**: Complete integration with Claude Desktop
- ✅ **Environment Variable Configuration**: Clean `ADB_PATH` and `IOS_WEBKIT_DEBUG_PROXY_PATH` support
- ✅ **Improved Error Handling**: Better diagnostics and path detection
- ✅ **No stdout Contamination**: Clean JSON-RPC communication
- ✅ **Cross-platform Support**: Windows, macOS, and Linux compatibility

## Features

- **MCP Server**: Integrates with Claude Desktop and other MCP-compatible AI assistants
- **Android Support**: Copy and restore Chrome tabs via ADB (Android Debug Bridge)
- **iOS Support**: Copy and restore Chrome/Safari tabs via iOS WebKit Debug Proxy
- **Cross-platform**: Works on Windows, macOS, and Linux
- **Standalone CLI**: Can be used independently without MCP integration

## MCP Integration

### Claude Desktop Setup

Add this configuration to your Claude Desktop config file (`~/Library/Application Support/Claude/claude_desktop_config.json` on macOS):

#### Option 1: Direct Path Configuration (Recommended)
```json
{
  "mcpServers": {
    "android-chrome": {
      "command": "/path/to/mcp-android-chrome",
      "args": ["mcp"],
      "env": {
        "ADB_PATH": "/Users/yourusername/Library/Android/sdk/platform-tools/adb",
        "IOS_WEBKIT_DEBUG_PROXY_PATH": "/opt/homebrew/bin/ios_webkit_debug_proxy"
      }
    }
  }
}
```

**To find your tool paths:**
- ADB: `which adb`
- iOS WebKit Debug Proxy: `which ios_webkit_debug_proxy`

#### Option 2: Auto-detection
```json
{
  "mcpServers": {
    "android-chrome": {
      "command": "/path/to/mcp-android-chrome",
      "args": ["mcp"]
    }
  }
}
```

The application will automatically detect common installation paths for ADB and iOS WebKit Debug Proxy.

### Available MCP Tools

- **`copy_tabs_android`**: Copy Chrome tabs from Android device via ADB
- **`copy_tabs_ios`**: Copy Chrome/Safari tabs from iOS device via WebKit Debug Proxy  
- **`reopen_tabs`**: Restore saved tabs to mobile devices
- **`check_environment`**: Verify system dependencies

### Available MCP Resources

- **`tabs://current`**: Access to currently loaded tabs (JSON format)

## Requirements

### For Android Support
- [Android Debug Bridge (ADB)](https://developer.android.com/studio/command-line/adb)
- Android device with USB debugging enabled
- Chrome browser running on Android device

### For iOS Support
- [iOS WebKit Debug Proxy](https://github.com/google/ios-webkit-debug-proxy)
- iOS device with Web Inspector enabled in Safari settings
- Chrome with Web Inspector enabled (iOS 16.4+) or Safari

### Installation

#### macOS (via Homebrew)
```bash
brew install --cask android-platform-tools
brew install ios-webkit-debug-proxy
```

#### Linux
```bash
sudo apt install android-tools-adb
# For iOS WebKit Debug Proxy, see: https://github.com/google/ios-webkit-debug-proxy
```

#### Windows
- Download Android Platform Tools from [developer.android.com](https://developer.android.com/tools/releases/platform-tools)
- iOS WebKit Debug Proxy is not officially supported on Windows

## Usage

### MCP Server Mode
```bash
# Start MCP server (for use with Claude Desktop)
mcp-android-chrome mcp
```

### Standalone CLI Mode

#### Copy tabs from Android
```bash
mcp-android-chrome android --port 9222 --debug
```

#### Copy tabs from iOS
```bash
mcp-android-chrome ios --port 9222 --debug
```

#### Restore tabs to device
```bash
# Save tabs to file first (copy output from android/ios commands)
echo '[{"id":"1","title":"Example","url":"https://example.com"}]' > tabs.json

# Restore to Android
mcp-android-chrome reopen --platform android tabs.json

# Restore to iOS
mcp-android-chrome reopen --platform ios tabs.json
```

#### Check system dependencies
```bash
# Check all platforms
mcp-android-chrome check

# Check specific platform
mcp-android-chrome check android
mcp-android-chrome check ios
```

## Device Setup

### Android Setup
1. Enable Developer Options in Android Settings
2. Enable USB Debugging option
3. Connect device via USB
4. Allow USB debugging when prompted
5. Start Chrome browser on device

### iOS Setup
1. Enable Safari Web Inspector in Settings > Safari > Advanced
2. For Chrome: Enable Web Inspector in Chrome Settings > Privacy and Security > Site Settings
3. Connect device via USB
4. Trust the computer when prompted
5. Start Chrome or Safari on device

## Development

### Building from Source
```bash
go mod tidy
go build -o mcp-android-chrome .
```

### Project Structure
```
mcp-android-chrome/
├── cmd/                 # CLI commands
├── internal/
│   ├── driver/         # Device drivers (Android/iOS)
│   ├── loader/         # HTTP/WebSocket communication
│   ├── mcp/           # MCP server implementation
│   ├── platform/      # OS utilities and dependency checking
│   └── template/      # HTML template generation
├── main.go
└── go.mod
```

## Technical Details

### How it works

#### Android
- Uses ADB to create port forwarding: `adb forward tcp:PORT localabstract:chrome_devtools_remote`
- Communicates with Chrome via [Chrome DevTools Protocol](https://chromedevtools.github.io/devtools-protocol/)
- Retrieves tabs via HTTP GET to `/json/list`
- Restores tabs via HTTP PUT to `/json/new?URL`

#### iOS
- Runs `ios_webkit_debug_proxy` as background process
- Communicates via [WebKit Inspector Protocol](https://github.com/WebKit/webkit/tree/main/Source/JavaScriptCore/inspector/protocol)
- Retrieves tabs via HTTP GET to `/json`
- Restores tabs via WebSocket connection with JavaScript injection

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Credits

- **Original tool**: [tab-transfer](https://github.com/machinateur/tab-transfer) by **machinateur**
- **Go port and MCP integration**: **kazuph**
- **MCP (Model Context Protocol)**: [Anthropic](https://github.com/anthropics/mcp)

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## Support

For issues related to:
- **Original PHP implementation**: See [machinateur/tab-transfer](https://github.com/machinateur/tab-transfer)
- **Go port and MCP features**: Open an issue in this repository
- **MCP protocol**: See [Anthropic MCP documentation](https://docs.anthropic.com/en/docs/build-with-claude/computer-use)