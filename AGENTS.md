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
- Resolve Codex user hooks through `$CODEX_HOME/hooks.json` with `~/.codex/hooks.json` as the default and canonicalize configured existing directories. Preserve additive TOML, project, plugin, and managed sources. Register exact `PreToolUse` and `PostToolUse` match-all command groups, retain the native payload, propagate freeze and revocation through exit code `2`, keep guard lookup and audit delivery fail-open with observable errors, and commit user hooks plus all four runtime artifacts through one rollback-capable transaction. Require `/hooks` approval for both definition hashes.
- Write Cursor hooks only to `~/.cursor/hooks.json`; keep project and enterprise sources read-only. Preserve user hooks, migrate the prior versionless Elydora contract, audit successful and failed tool calls, emit valid native JSON responses, retain PowerShell exit codes, and commit guard, runtime metadata, private key, audit runtime, and user hooks in one rollback-capable transaction.
- Model stable, legacy, and early-access hook generations as explicit contracts. Keep their activation requirements visible in CLI output and README guidance.
- Select Kimi Code from an explicit `KIMI_CODE_HOME` or an existing `~/.kimi-code` home, and select legacy `kimi-cli` from an existing `~/.kimi` home. An empty `KIMI_CODE_HOME` uses `~/.kimi-code`; executable lookup is outside the activation contract, and equal stable and legacy config paths collapse to the stable contract.
- Preserve Kimi TOML comments and unrelated formatting through range-based edits, then parse the complete rendered document before writing it. Own only exact three-field hooks for `PreToolUse`, `PostToolUse`, and `PostToolUseFailure`; require one complete triple per runtime identity, use encoded PowerShell commands on Windows, and retain the native snake_case payload.
- Generate Kimi guard, config, private key, and audit runtimes together with every selected TOML document in one rollback-capable transaction. Preflight physical paths and existing runtime identity before creating state; status validates the complete runtime metadata, canonical private key, exact agent identity, and non-empty physical scripts.
- Keep Grok Build writes inside its native global `$GROK_HOME/hooks/*.json` contract. Treat project `.grok/hooks`, plugin hooks, Claude Code, and Cursor compatibility files as read-only integration sources. Own only exact matchless three-field command groups for `PreToolUse`, `PostToolUse`, and `PostToolUseFailure`; accept official non-negative integer timeouts and string-valued environments; reject lifecycle matchers; and preserve native camelCase payloads.
- Generate Grok guard, config, private key, and audit runtimes with the user hook document in one rollback-capable transaction. Use encoded PowerShell commands on Windows, emit native deny JSON with exit code `2`, preflight physical paths and runtime identity, and require one exact event triple plus strict runtime metadata and a canonical private key for healthy status.
- Write Auggie hooks only to `~/.augment/settings.json`; keep system and workspace settings read-only. Generate `.cmd` wrappers on Windows and `.sh` wrappers on Unix because Auggie dispatches supported script paths, and express hook timeouts in milliseconds.
- Write Cline hooks only to `$CLINE_DIR/hooks` with `~/.cline/hooks` as the default; keep Documents and workspace hook roots read-only. Preserve official input byte-for-byte and translate guard exit code `2` into Cline's JSON stdout cancellation control.
- Write GitHub Copilot CLI hooks only to `$COPILOT_HOME/hooks/elydora-audit.json` with `~/.copilot/hooks/elydora-audit.json` as the default. Migrate exact Elydora-owned entries from project `.github/hooks/hooks.json`, preserve native camelCase input, require `disableAllHooks` to be false, and commit runtime plus hook documents in one rollback-capable transaction.
- Factory Droid `droid@0.175.0` reads direct event maps from `~/.factory/hooks.json` or the legacy `~/.factory/hooks/hooks.json`; `~/.factory/settings.json` stores the same map under `hooks`. Select the active source per event, preserve JSONC syntax, and direct users to `/hooks` for review.
- Commit Factory Droid runtime config, private key, audit runtime, and affected hook documents in one rollback-capable transaction. Status requires an enabled `PreToolUse`/`PostToolUse` pair with matching managed runtime files.
- Qwen Code `0.20.0` loads user hooks from `$QWEN_HOME/settings.json` with `~/.qwen/settings.json` as the default; follow Qwen's `.env` precedence and keep workspace settings read-only. Preserve comments, reject trailing commas and duplicate keys, express timeouts in milliseconds, and direct users to `/hooks` for review.
- Commit Qwen Code runtime config, private key, audit runtime, and user settings in one rollback-capable transaction. Status requires enabled `PreToolUse` and `PostToolUse` hooks with matching managed runtime files.

## Code Quality

- Preserve the minimum Go version declared in `go.mod`.
- Keep production source files at or below 500 lines.
- Keep functions focused on one ownership boundary.
- Return unexpected errors to the CLI boundary with operation and path context.
- Use documented defaults only for genuinely optional configuration.
- Avoid compatibility shims without a named public or user configuration contract.
- Resolve every agent runtime directory as one physical child of `~/.elydora`; reject separators, traversal segments, cross-platform reserved names, symbolic-link directories, and linked identity configs before writes or recursive removal. Validate stored directory identity before changing host CLI configuration, and require an explicit agent ID when discovery is ambiguous.
- Accept CLI install credentials through terminal-echo-disabled input or physical owner-only credential files. Keep credentials out of process arguments, reject legacy secret options with redacted errors, and require one UTF-8 line of at most 64 KiB.
- Commit runtime config, private key, and audit script through one rollback-capable transaction. Persist the signing key once with mode `0600`; generated runtimes validate physical identity, size, and Unix permissions before reading config or key material.
- Write the guard runtime through a flushed same-directory temporary file followed by atomic replacement. Surface permission, close, synchronization, commit, cleanup, and recovery failures.
- Read status cache, chain state, and error logs through physical descriptors with identity checks. Write cache and validated chain state atomically, and append error logs through no-follow owner-only descriptors. Preserve rollback artifacts when recovery cannot safely restore an original file and include the recovery path in the surfaced error.

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
