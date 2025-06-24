package template

import (
	"fmt"
	"os"
)

// WebSocketClientTemplate generates HTML files with embedded WebSocket clients
type WebSocketClientTemplate struct{}

// NewWebSocketClientTemplate creates a new WebSocket client template generator
func NewWebSocketClientTemplate() *WebSocketClientTemplate {
	return &WebSocketClientTemplate{}
}

// Generate creates HTML content with embedded WebSocket client for tab restoration
func (w *WebSocketClientTemplate) Generate(tabsJSON, wsURL string, debug bool) string {
	debugScript := ""
	if debug {
		debugScript = `
			console.log('Debug mode enabled');
			ws.onmessage = function(event) {
				console.log('WebSocket response:', event.data);
			};`
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>iOS Tab Restorer - MCP Android Chrome</title>
    <meta charset="UTF-8">
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            max-width: 800px;
            margin: 0 auto;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .container {
            background: white;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        .status {
            padding: 10px;
            margin: 10px 0;
            border-radius: 4px;
            font-weight: bold;
        }
        .success { background-color: #d4edda; color: #155724; border: 1px solid #c3e6cb; }
        .error { background-color: #f8d7da; color: #721c24; border: 1px solid #f5c6cb; }
        .info { background-color: #d1ecf1; color: #0c5460; border: 1px solid #bee5eb; }
        .progress { background-color: #fff3cd; color: #856404; border: 1px solid #ffeaa7; }
        .tab-list {
            max-height: 300px;
            overflow-y: auto;
            border: 1px solid #ddd;
            padding: 10px;
            margin: 10px 0;
            background-color: #f9f9f9;
        }
        .tab-item {
            padding: 5px;
            border-bottom: 1px solid #eee;
            font-size: 14px;
        }
        .tab-item:last-child { border-bottom: none; }
        .tab-title { font-weight: bold; }
        .tab-url { color: #666; font-size: 12px; }
        button {
            background-color: #007bff;
            color: white;
            border: none;
            padding: 10px 20px;
            border-radius: 4px;
            cursor: pointer;
            font-size: 16px;
            margin: 10px 5px;
        }
        button:hover { background-color: #0056b3; }
        button:disabled { background-color: #6c757d; cursor: not-allowed; }
        .footer {
            margin-top: 20px;
            padding-top: 20px;
            border-top: 1px solid #ddd;
            font-size: 12px;
            color: #666;
            text-align: center;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>iOS Tab Restorer</h1>
        <p>This tool will restore tabs to your iOS device using WebKit Debug Protocol.</p>
        
        <div id="status" class="status info">
            Ready to restore tabs. Click "Start Restoration" to begin.
        </div>

        <div class="tab-list">
            <h3>Tabs to restore:</h3>
            <div id="tab-list"></div>
        </div>

        <button id="startBtn" onclick="startRestoration()">Start Restoration</button>
        <button id="cancelBtn" onclick="cancelRestoration()" disabled>Cancel</button>

        <div class="footer">
            <p>iOS Tab Restorer - Part of MCP Android Chrome</p>
            <p>Original tool by machinateur, Go port by kazuph</p>
        </div>
    </div>

    <script>
        const tabs = %s;
        const wsURL = '%s';
        let ws = null;
        let isRestoring = false;
        let restoredCount = 0;

        // Display tabs to be restored
        function displayTabs() {
            const tabListEl = document.getElementById('tab-list');
            tabListEl.innerHTML = '';
            
            tabs.forEach((tab, index) => {
                const tabEl = document.createElement('div');
                tabEl.className = 'tab-item';
                tabEl.innerHTML = ` + "`" + `
                    <div class="tab-title">${index + 1}. ${tab.title || 'Untitled'}</div>
                    <div class="tab-url">${tab.url}</div>
                ` + "`" + `;
                tabListEl.appendChild(tabEl);
            });
        }

        function updateStatus(message, type = 'info') {
            const statusEl = document.getElementById('status');
            statusEl.textContent = message;
            statusEl.className = 'status ' + type;
        }

        function startRestoration() {
            if (isRestoring) return;
            
            isRestoring = true;
            restoredCount = 0;
            
            document.getElementById('startBtn').disabled = true;
            document.getElementById('cancelBtn').disabled = false;
            
            updateStatus('Connecting to WebKit Debug Proxy...', 'progress');
            
            try {
                ws = new WebSocket(wsURL);
                
                ws.onopen = function() {
                    updateStatus('Connected! Starting tab restoration...', 'progress');
                    restoreNextTab();
                };
                
                ws.onerror = function(error) {
                    updateStatus('WebSocket connection failed. Make sure iOS WebKit Debug Proxy is running.', 'error');
                    resetButtons();
                };
                
                ws.onclose = function() {
                    if (isRestoring) {
                        updateStatus('Connection closed unexpectedly.', 'error');
                        resetButtons();
                    }
                };
                
                %s
                
            } catch (error) {
                updateStatus('Failed to connect: ' + error.message, 'error');
                resetButtons();
            }
        }

        function restoreNextTab() {
            if (restoredCount >= tabs.length) {
                updateStatus(` + "`" + `Successfully restored ${restoredCount} tabs!` + "`" + `, 'success');
                finishRestoration();
                return;
            }
            
            const tab = tabs[restoredCount];
            updateStatus(` + "`" + `Restoring tab ${restoredCount + 1}/${tabs.length}: ${tab.title}` + "`" + `, 'progress');
            
            // Send WebKit Debug Protocol message to open new tab
            const message = {
                id: restoredCount + 1,
                method: 'Runtime.evaluate',
                params: {
                    expression: ` + "`" + `window.open('${tab.url.replace(/'/g, "\\'")}');` + "`" + `
                }
            };
            
            ws.send(JSON.stringify(message));
            
            restoredCount++;
            
            // Continue with next tab after a short delay
            setTimeout(restoreNextTab, 300);
        }

        function cancelRestoration() {
            if (ws) {
                ws.close();
            }
            updateStatus('Restoration cancelled.', 'info');
            resetButtons();
        }

        function finishRestoration() {
            if (ws) {
                ws.close();
            }
            resetButtons();
        }

        function resetButtons() {
            isRestoring = false;
            document.getElementById('startBtn').disabled = false;
            document.getElementById('cancelBtn').disabled = true;
        }

        // Initialize the page
        displayTabs();
        
        // Auto-start restoration after a short delay (optional)
        // setTimeout(startRestoration, 2000);
    </script>
</body>
</html>`, tabsJSON, wsURL, debugScript)
}

// WriteFile writes content to a file
func WriteFile(filepath, content string) error {
	return os.WriteFile(filepath, []byte(content), 0644)
}