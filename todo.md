# TODO List - MCP Android Chrome

## v0.9.1 - Enhanced Tab Management Features

### 🔄 Current Tabs Resource Enhancement
- [x] **Auto-populate current_tabs on startup**
  - [x] Implement automatic Android tab fetching on MCP server startup
  - [x] Cache latest 30 tabs (URL + Title) without user confirmation
  - [x] Update current_tabs resource with real data instead of empty array
  - [x] Add configuration option for cache size (default: 30)
  - [x] Handle connection failures gracefully
  - [x] **FIXED**: Added cache refresh tool for manual updates
  - [x] **FIXED**: Added cache status information with timestamps
  - [x] **FIXED**: Added RefreshTabCacheArgs and CacheStatusArgs types
  - [x] **VERIFIED**: Manual cache refresh works (caches 30 tabs)
  - [ ] **REMAINING**: Test resource reading after cache population in Claude Desktop
  - [ ] **ENHANCEMENT**: Consider adding cache auto-refresh interval

### 📄 YAML Format Support
- [x] **Add YAML output format option**
  - [x] Add format parameter to all tab-related tools (copy_tabs_android, copy_tabs_ios)
  - [x] Support both JSON and YAML output formats via format.Formatter
  - [x] Default to JSON for backward compatibility
  - [x] Add YAML formatting for current_tabs resource (tabs://current-yaml)
  - [x] Created internal/format package with TabFormatter
  - [x] Support "yaml", "yml", "json" format parameters
  - [x] **VERIFIED**: YAML output works perfectly via MCP tools
  - [ ] **ENHANCEMENT**: Update CLI tools to support --format yaml flag

### 🗑️ Tab Closing Features
- [ ] **Single tab closing**
  - [ ] Implement `close_tab` MCP tool
  - [ ] Support tab closing by Chrome tab ID
  - [ ] Add confirmation option for safety
  
- [ ] **Bulk tab closing**
  - [ ] Implement `close_tabs_bulk` MCP tool  
  - [ ] Support multiple tab IDs in single operation
  - [ ] Add filtering capabilities (by URL pattern, title keywords, etc.)
  - [ ] Integration with Claude for natural language tab selection

- [ ] **Tab ID Investigation**
  - [ ] Research Chrome DevTools Protocol tab ID system
  - [ ] Verify if Chrome assigns unique IDs to tabs
  - [ ] Document tab ID lifecycle and stability
  - [ ] Test tab ID persistence across browser restarts

### 🔍 Tab Search Features
- [ ] **Search functionality assessment**
  - [ ] Review current search capabilities in existing tools
  - [ ] Implement dedicated `search_tabs` MCP tool if needed
  - [ ] Support search by URL, title, domain patterns
  - [ ] Add fuzzy search capabilities
  - [ ] Return ranked results with relevance scores

### 🏗️ Architecture Improvements
- [ ] **Background tab monitoring**
  - [ ] Implement periodic tab cache refresh
  - [ ] Add tab change detection (new/closed tabs)
  - [ ] Optimize network calls to avoid excessive requests
  
- [ ] **Error handling enhancements**
  - [ ] Improve error messages for tab operations
  - [ ] Add retry logic for failed operations
  - [ ] Better handling of device disconnection scenarios

### 📚 Documentation Updates
- [ ] Update README with new tab management features
- [ ] Add examples for Claude Desktop integration
- [ ] Document tab filtering and search patterns
- [ ] Create troubleshooting guide for common issues

## Research Questions

1. **Chrome DevTools Protocol Tab IDs**: 
   - Do tabs have persistent unique IDs?
   - What's the format and lifecycle of tab IDs?
   - Can we rely on these IDs for bulk operations?

2. **Tab Closing Mechanisms**:
   - Does CDP support `/json/close/{id}` endpoint?
   - What happens when closing non-existent tabs?
   - Are there rate limits for bulk operations?

3. **Background Operations**:
   - Should we run background polling for tab changes?
   - How to balance freshness vs. performance?
   - Memory management for cached tab data?

## Implementation Priority

1. 🥇 **Priority 1**: Auto-populate current_tabs (user requested)
2. 🥈 **Priority 2**: Single tab closing functionality  
3. 🥉 **Priority 3**: Bulk tab closing with filtering
4. 🏅 **Priority 4**: Enhanced search capabilities
5. 🎯 **Priority 5**: Background monitoring and caching