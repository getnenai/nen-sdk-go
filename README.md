# nen-sdk-go

Go SDK for the [Nen Desktop API](https://getnen.ai). Create cloud desktops, execute computer-use tools, and manage RDP sessions programmatically.

## Installation

```bash
go get github.com/getnenai/nen-sdk-go
```

## Quick Start

```go
package main

import (
	"context"
	"fmt"

	nendesktop "github.com/getnenai/nen-sdk-go"
)

func main() {
	client := nendesktop.New(nendesktop.Config{APIKey: "sk_nen_..."})
	ctx := context.Background()

	// Create a desktop
	desktop, _ := client.CreateDesktop(ctx, "sandbox")
	fmt.Printf("Created: %s (status: %s)\n", desktop.DesktopID, desktop.Status)

	// Check its status
	desktop, _ = client.GetDesktop(ctx, desktop.DesktopID)
	fmt.Printf("Status: %s, IP: %s\n", desktop.Status, desktop.PublicIP)

	// Clean up
	client.DeleteDesktop(ctx, desktop.DesktopID)
}
```

## Configuration

```go
client := nendesktop.New(nendesktop.Config{
	APIKey:  "sk_nen_...",
	BaseURL: "https://desktop.api.getnen.ai", // default
	Timeout: 30 * time.Second,                // default
})
```

The `Execute` method uses a 120-second timeout regardless of the client timeout, since tool execution can be slow.

## API Reference

### Desktop CRUD

| Method | Description |
|--------|-------------|
| `CreateDesktop(ctx, desktopType)` | Create a new desktop. Returns `*Desktop`. |
| `ListDesktops(ctx)` | List all active desktops. Returns `[]Desktop`. |
| `GetDesktop(ctx, desktopID)` | Get a single desktop. Returns `*Desktop`. |
| `UpdateDesktop(ctx, desktopID, name)` | Update desktop name. Returns `*Desktop`. |
| `DeleteDesktop(ctx, desktopID)` | Delete a desktop. Returns `*DeleteResponse`. |

### Tool Execution

| Method | Description |
|--------|-------------|
| `Execute(ctx, desktopID, tool, action, params)` | Execute a tool action. Returns `*ExecuteResult`. |
| `ListTools(ctx, desktopID)` | List available tools. Returns `[]ToolSchema`. |
| `GetToolLogs(ctx, desktopID)` | Get tool execution logs. Returns `json.RawMessage`. |

### Sessions

| Method | Description |
|--------|-------------|
| `CreateSession(ctx, desktopID)` | Create or reconnect an RDP session. Returns `*SessionInfo`. |
| `GetSession(ctx, desktopID)` | Get session status. Returns `*SessionInfo`. |
| `DeleteSession(ctx, desktopID)` | Disconnect the session. Returns `error`. |

## Error Handling

All API errors return `*APIError`, which carries `StatusCode` and `Body`:

```go
desktop, err := client.GetDesktop(ctx, "dsk_nonexistent")
var apiErr *nendesktop.APIError
if errors.As(err, &apiErr) {
	fmt.Printf("Error %d: %s\n", apiErr.StatusCode, apiErr.Body)
}
```

## Zero Dependencies

This SDK uses only the Go standard library.

## Testing

Run the E2E test suite against the production API:

```bash
export NEN_API_KEY=sk_nen_...
make -C sdks test-go
```

Tests require a valid API key with desktop create/delete permissions.
Override the API host: `make -C sdks test-go NEN_API_URL=https://custom-host`

## Examples

See the full agent example in [`cmd/nen/templates/anthropic-computer-use/agent.py`](../../cmd/nen/templates/anthropic-computer-use/agent.py).
