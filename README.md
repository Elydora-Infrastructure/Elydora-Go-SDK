# Elydora Go SDK

Official Go SDK for the [Elydora](https://elydora.com) tamper-evident audit platform. Build cryptographically verifiable audit trails for AI agent operations.

## Installation

```bash
go get github.com/Elydora-Infrastructure/Elydora-Go-SDK
```

Requires Go 1.21+. Zero third-party dependencies (stdlib only).

## Quick Start

```go
package main

import (
	"fmt"
	"log"

	elydora "github.com/Elydora-Infrastructure/Elydora-Go-SDK"
)

func main() {
	// Authenticate
	auth, err := elydora.Login("https://api.elydora.com", "user@example.com", "password")
	if err != nil {
		log.Fatal(err)
	}

	// Create client
	client, err := elydora.NewClient(&elydora.Config{
		OrgID:      auth.User.OrgID,
		AgentID:    "my-agent-id",
		PrivateKey: "<base64url-encoded-ed25519-seed>",
		Token:      auth.Token,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Create and submit an operation
	eor, err := client.CreateOperation(&elydora.CreateOperationParams{
		OperationType: "data.access",
		Subject:       map[string]interface{}{"user_id": "u-123"},
		Action:        map[string]interface{}{"type": "read"},
		Payload:       map[string]interface{}{"record_id": "rec-456"},
	})
	if err != nil {
		log.Fatal(err)
	}

	resp, err := client.SubmitOperation(eor)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Receipt: %s\n", resp.Receipt.ReceiptID)
}
```

## CLI

The SDK includes a CLI for installing audit hooks into AI coding agents.

```bash
go install github.com/Elydora-Infrastructure/Elydora-Go-SDK/cmd/elydora@latest

elydora install \
  --agent claudecode \
  --org-id org-123 \
  --agent-id agent-456 \
  --private-key <key> \
  --kid agent-456-key-v1
```

### Commands

| Command | Description |
|---------|-------------|
| `elydora install` | Install Elydora audit hook for a coding agent |
| `elydora uninstall` | Remove Elydora audit hook for a coding agent |
| `elydora status` | Show installation status for all agents |
| `elydora agents` | List supported coding agents |

### Supported Agents

| Agent | Key |
|-------|-----|
| Claude Code | `claudecode` |
| Cursor | `cursor` |
| Gemini CLI | `gemini` |
| Augment Code | `augment` |
| Kiro | `kiro` |
| OpenCode | `opencode` |

## API Reference

### Configuration

```go
client, err := elydora.NewClient(&elydora.Config{
	OrgID:      "org-123",       // Organization ID
	AgentID:    "agent-456",     // Agent ID
	PrivateKey: "<seed>",        // Base64url-encoded Ed25519 seed
	BaseURL:    "https://...",   // API base URL (default: https://api.elydora.com)
	TTLMs:      30000,           // Operation TTL in ms (default: 30000)
	MaxRetries: 3,               // Max retries on transient failures (default: 3)
	Token:      "<jwt>",         // JWT token for authenticated requests
})
```

### Authentication

```go
// Register a new user and organization
reg, err := elydora.Register(baseURL, email, password,
	elydora.WithDisplayName("Alice"),
	elydora.WithOrgName("Acme Corp"),
)

// Login and receive a JWT
auth, err := elydora.Login(baseURL, email, password)
```

### Operations

```go
// Create a signed EOR locally (no network call)
eor, err := client.CreateOperation(&elydora.CreateOperationParams{
	OperationType: "inference",
	Subject:       map[string]interface{}{"model": "gpt-4"},
	Action:        map[string]interface{}{"type": "completion"},
	Payload:       map[string]interface{}{"prompt": "Hello"},
	KID:           "agent-456-key-v1",
})

// Submit to API
resp, err := client.SubmitOperation(eor)

// Retrieve an operation
op, err := client.GetOperation(operationID)

// Verify integrity
result, err := client.VerifyOperation(operationID)
```

### Agent Management

```go
// Register a new agent
agent, err := client.RegisterAgent(&elydora.RegisterAgentRequest{
	AgentID:           "my-agent",
	DisplayName:       "My Agent",
	ResponsibleEntity: "team@example.com",
	Keys: []elydora.AgentKeyRequest{
		{KID: "key-v1", PublicKey: "<base64url>", Algorithm: "ed25519"},
	},
})

// Get agent details
details, err := client.GetAgent(agentID)

// Freeze an agent
err := client.FreezeAgent(agentID, "security review")

// Revoke a key
err := client.RevokeKey(agentID, kid, "key rotation")
```

### Audit

```go
results, err := client.QueryAudit(&elydora.AuditQueryRequest{
	AgentID:       "agent-123",
	OperationType: "inference",
	StartTime:     &startTime,
	EndTime:       &endTime,
	Limit:         &limit,
})
```

### Epochs

```go
epochs, err := client.ListEpochs()
epoch, err := client.GetEpoch(epochID)
```

### Exports

```go
export, err := client.CreateExport(&elydora.CreateExportRequest{
	StartTime: startTime,
	EndTime:   endTime,
	Format:    "json",
})

exports, err := client.ListExports()
detail, err := client.GetExport(exportID)
```

### JWKS

```go
jwks, err := client.GetJWKS()
```

## Error Handling

```go
import "errors"

var apiErr *elydora.ElydoraError
if errors.As(err, &apiErr) {
	fmt.Println(apiErr.Code)       // e.g. "INVALID_SIGNATURE"
	fmt.Println(apiErr.Message)    // Human-readable message
	fmt.Println(apiErr.StatusCode) // HTTP status code
	fmt.Println(apiErr.RequestID)  // Request ID for support
}
```

## License

MIT
