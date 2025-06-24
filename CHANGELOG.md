# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.9.0] - 2025-06-24

### Added
- **MCP Server Implementation**: Complete Model Context Protocol server for Claude Desktop integration
- **Environment Variable Configuration**: Support for `ADB_PATH` and `IOS_WEBKIT_DEBUG_PROXY_PATH` environment variables
- **Automatic Path Detection**: Fallback to common installation paths for ADB and iOS WebKit Debug Proxy
- **MCP Tools**: Four main tools available via MCP interface:
  - `copy_tabs_android`: Copy Chrome tabs from Android device
  - `copy_tabs_ios`: Copy Chrome/Safari tabs from iOS device  
  - `reopen_tabs`: Restore tabs to mobile devices
  - `check_environment`: Verify system dependencies
- **MCP Resources**: `tabs://current` resource for accessing loaded tabs
- **Cross-platform Support**: Windows, macOS, and Linux compatibility
- **Comprehensive Documentation**: Setup guides for Claude Desktop integration

### Fixed
- **stdout Contamination**: All debug output redirected to stderr to prevent JSON-RPC interference
- **JSON Schema Truncation**: Shortened parameter descriptions to prevent schema truncation
- **Path Resolution**: Improved ADB and iOS WebKit Debug Proxy path detection
- **Error Diagnostics**: Enhanced error messages with specific installation instructions

### Technical
- **Go Implementation**: Complete rewrite of PHP original in Go
- **Cobra CLI Framework**: Modern command-line interface
- **metoro-io/mcp-golang**: MCP server implementation using official Go library
- **Gorilla WebSocket**: WebSocket communication for iOS tab restoration
- **Chrome DevTools Protocol**: Android tab management via CDP
- **WebKit Inspector Protocol**: iOS tab management via WebKit protocol

### Dependencies
- Go 1.21+
- Android Debug Bridge (ADB) for Android support
- iOS WebKit Debug Proxy for iOS support

## Credits

- **Original Implementation**: [machinateur/tab-transfer](https://github.com/machinateur/tab-transfer)
- **Go Port and MCP Integration**: kazuph
- **Model Context Protocol**: [Anthropic](https://github.com/anthropics/mcp)