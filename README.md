# Elydora Go SDK

Official Go SDK for the Elydora AI audit trail platform.

## Installation

```bash
go get github.com/elydora/sdk-go
```

## Quick Start

```go
package main

import (
	"fmt"
	"log"

	elydora "github.com/elydora/sdk-go"
)

func main() {
	// Login
	auth, err := elydora.Login("https://api.elydora.com", "user@example.com", "password")
	if err != nil {
		log.Fatal(err)
	}

	// Create client
	client, err := elydora.NewClient(&elydora.Config{
		OrgID:      "org-123",
		AgentID:    "agent-456",
		PrivateKey: "base64url-encoded-ed25519-seed",
		Token:      auth.Token,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Create and submit an operation
	eor, err := client.CreateOperation(&elydora.CreateOperationParams{
		OperationType: "inference",
		Subject:       map[string]interface{}{"model": "gpt-4"},
		Action:        map[string]interface{}{"type": "completion"},
		Payload:       map[string]interface{}{"prompt": "Hello"},
		KID:           "agent-456-key-1",
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

## Requirements

- Go 1.21 or later
- No third-party dependencies (stdlib only)
