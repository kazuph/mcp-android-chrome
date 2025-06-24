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
- [x] **Single tab closing**
  - [x] Implement `close_tab` MCP tool
  - [x] Support tab closing by Chrome tab ID via /json/close/{id} endpoint
  - [x] Add confirmation option for safety (confirm=true required)
  - [x] **VERIFIED**: Tool registered and accepts tabId parameter
  
- [x] **Bulk tab closing**
  - [x] Implement `close_tabs_bulk` MCP tool  
  - [x] Support multiple tab IDs in single operation
  - [x] Add filtering capabilities (by URL pattern, title keywords, etc.)
  - [x] Integration with Claude for natural language tab selection via filters
  - [x] **SAFETY**: Added dryRun option to preview before closing
  - [x] **SAFETY**: Confirmation required (confirm=true) for actual closing

- [x] **Tab ID Investigation**
  - [x] Research Chrome DevTools Protocol tab ID system (/json/close/{id})
  - [x] Verify Chrome assigns unique IDs (confirmed: "11952", "C4590D171DDF33989C7B1ED6DFE754FC")
  - [x] Implement HTTP POST to /json/close/{id} endpoint
  - [ ] **REMAINING**: Test actual tab closing functionality with real device
  - [ ] **REMAINING**: Document any edge cases or limitations found

### 🔍 Tab Search Features
- [x] **Search functionality assessment**
  - [x] Review current search capabilities in existing tools
  - [x] Implement dedicated `search_tabs` MCP tool
  - [x] Support search by URL, title, domain patterns
  - [x] Add fuzzy search capabilities with relevance scoring
  - [x] Return ranked results with relevance scores
  - [x] **IMPLEMENTED**: Added `search_tabs` MCP tool with advanced filtering
  - [x] **FEATURES**: Query, domain, title, URL filters with relevance scoring
  - [x] **FORMATS**: Support for both JSON and YAML output formats
  - [x] **RANKING**: Smart relevance scoring with exact match bonuses

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