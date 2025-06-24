# Claude Desktop Setup Guide

## Configuration File Location

**macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`

## Recommended Configuration

### Option 1: Basic Configuration (Recommended)
The application now automatically detects ADB and iOS WebKit Debug Proxy paths, so a basic configuration should work:

```json
{
  "mcpServers": {
    "android-chrome": {
      "command": "/Users/kazuph/src/github.com/machinateur/tab-transfer/mcp-android-chrome",
      "args": ["mcp"]
    }
  }
}
```

### Option 2: Explicit PATH Configuration (If Needed)
If you encounter PATH-related issues, you can explicitly set the PATH environment variable:

```json
{
  "mcpServers": {
    "android-chrome": {
      "command": "/Users/kazuph/src/github.com/machinateur/tab-transfer/mcp-android-chrome",
      "args": ["mcp"],
      "env": {
        "PATH": "/Users/kazuph/Library/Android/sdk/platform-tools:/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin"
      }
    }
  }
}
```

### Option 3: Full Environment Configuration
For maximum compatibility, you can set multiple environment variables:

```json
{
  "mcpServers": {
    "android-chrome": {
      "command": "/Users/kazuph/src/github.com/machinateur/tab-transfer/mcp-android-chrome",
      "args": ["mcp"],
      "env": {
        "PATH": "/Users/kazuph/Library/Android/sdk/platform-tools:/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin",
        "ANDROID_HOME": "/Users/kazuph/Library/Android/sdk",
        "ANDROID_SDK_ROOT": "/Users/kazuph/Library/Android/sdk"
      }
    }
  }
}
```

## Setup Steps

1. **Install dependencies** (if not already installed):
   ```bash
   # For Android support
   brew install --cask android-platform-tools
   
   # For iOS support  
   brew install ios-webkit-debug-proxy
   ```

2. **Build the MCP server** (if not already built):
   ```bash
   cd /Users/kazuph/src/github.com/machinateur/tab-transfer
   go build -o mcp-android-chrome .
   ```

3. **Create or edit Claude Desktop config**:
   ```bash
   # Create config directory if it doesn't exist
   mkdir -p "~/Library/Application Support/Claude"
   
   # Edit config file
   open "~/Library/Application Support/Claude/claude_desktop_config.json"
   ```

4. **Add the configuration** (use Option 1 first):
   - Copy the JSON configuration above
   - Replace `/Users/kazuph/src/github.com/machinateur/tab-transfer/mcp-android-chrome` with the actual path to your binary

5. **Restart Claude Desktop**

6. **Test the connection**:
   - Open Claude Desktop
   - Try using MCP tools like `copy_tabs_android` or `check_environment`

## Troubleshooting

### Common Issues

1. **"Command not found" errors**:
   - Use Option 2 or 3 configuration with explicit PATH
   - Verify binary path is correct
   - Check that dependencies are installed

2. **"Permission denied" errors**:
   - Make sure the binary is executable: `chmod +x mcp-android-chrome`
   - Check file permissions

3. **"Device not found" errors**:
   - Enable USB debugging on Android device
   - Enable Web Inspector on iOS device
   - Connect device via USB and authorize computer

### Debug Commands

```bash
# Test MCP server manually
./mcp-android-chrome check

# Check ADB devices
adb devices

# Check paths
which adb
which ios_webkit_debug_proxy
```

### Log Files

Check Claude Desktop logs for detailed error information:
- **macOS**: `~/Library/Logs/Claude/mcp-server-android-chrome.log`

## Available MCP Tools

Once configured, these tools will be available in Claude Desktop:

- **`copy_tabs_android`**: Copy tabs from Android Chrome
- **`copy_tabs_ios`**: Copy tabs from iOS Chrome/Safari  
- **`reopen_tabs`**: Restore tabs to mobile device
- **`check_environment`**: Verify system dependencies

## Security Notes

- The MCP server only runs when Claude Desktop is active
- No network communication - only local USB device access
- All data stays on your local machine
- USB debugging should be disabled when not in use