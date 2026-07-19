# Elydora Go SDK Engineering Contract

## Scope

This repository owns the public Go SDK, the `cmd/elydora` CLI, local signing behavior, generated Node.js hook runtimes, and Go-specific agent adapters.

## Integration Sources

- Verify every agent hook contract against current official provider documentation before changing an adapter.
- Treat `../Elydora-Open-Source/integrations/catalog.json` as the cross-product provider inventory.
- Keep adapter delivery claims aligned with executable tests in this repository.
- Synchronize completed provider behavior into the Node SDK, Python SDK, Open Source distribution, Console, Docs, and landing page through separate reviewed commits.

## Hook Adapter Invariants

- Preserve unrelated user configuration and remove only Elydora-owned entries.
- Parse every affected user configuration before the first write.
- Surface malformed or unreadable configuration with contextual errors and leave the original file intact.
- Write configuration through a same-directory temporary file, flush it, and atomically replace the target.
- Resolve and quote the Node.js runtime and generated script paths for the host shell.
- Forward official hook JSON from STDIN without reshaping provider fields.
- Use the provider's documented blocking mechanism. Command-hook providers that define exit code `2` must receive exit code `2` from the freeze guard.
- Report installation as healthy only when a complete hook contract references both generated runtime scripts and both scripts exist.
- Model stable, legacy, and early-access hook generations as explicit contracts. Keep their activation requirements visible in CLI output and README guidance.
- Select Kimi Code and legacy `kimi-cli` contracts from runtime evidence. An empty `KIMI_CODE_HOME` uses `~/.kimi-code`; create no cross-runtime migration marker.
- Preserve Kimi TOML comments and unrelated formatting through range-based edits, then parse the complete rendered document before writing it.

## Code Quality

- Preserve the minimum Go version declared in `go.mod`.
- Keep production source files at or below 500 lines.
- Keep functions focused on one ownership boundary.
- Return unexpected errors to the CLI boundary with operation and path context.
- Use documented defaults only for genuinely optional configuration.
- Avoid compatibility shims without a named public or user configuration contract.

## Verification

Run the focused adapter test during development, then execute all release gates before commit:

```powershell
go test ./cmd/elydora/plugins -run <Provider> -count=1
go test ./...
go test -race ./...
go vet ./...
go build ./cmd/elydora
govulncheck ./...
git diff --check
```

Provider adapter tests must cover installation, idempotency, official event forwarding, blocking behavior, status, missing runtime files, uninstall ownership, and malformed configuration preservation.

Commit and push one root issue before starting the next one.
