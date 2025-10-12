# Plan: Git-based ChatGPT (Wails + Go)

## FRONT-AND-CENTER PRINCIPLE — Every Message Is a Commit
- All chat turns (USER and AGENT) are persisted as Git commits on the active branch.
- Messages without file changes are committed with `git commit --allow-empty`.
- Commit messages MUST start with a role prefix: `USER:` or `AGENT:` followed by a concise summary; the full text goes in the body.
- Manual file edits are also committed and use the appropriate role prefix (typically `USER:` for direct edits; `AGENT:` when applying approved agent patches).
- If a message leads to code changes, those changes are a separate follow-up commit that references the message commit via a footer: `Relates-To: <message-sha>`.

### Commit Message Standard (required)
Subject: `USER: <summary>` or `AGENT: <summary>`

Body:
- Full message text (markdown allowed)
- Optional footers:
  - `Relates-To: <sha>` (link patch commit to the originating message commit)
  - `Session: <session-id>`
  - `Type: message|patch|manual-edit`

Examples:
- `USER: How do we handle side-thread expansion?` (empty commit)
- `AGENT: Explanation and plan for connectors` (empty commit)
- `AGENT: Apply connectors for merge commits` (patch commit; Body includes diff summary; Footer `Relates-To: <sha-of-agent-message>`)
- `USER: Manual edit: tweak layout breakpoints` (patch commit for direct edit)

## Goals
- Make the Git history the chat itself: EVERY message is a commit (empty commits allowed), clearly marked by `USER:`/`AGENT:` prefixes.
- Desktop app that treats a Git repo like a chat: main branch as the narrative spine, merges as expandable side threads, commits/messages as bubbles with long text.
- Let the AI propose patches, preview diffs, apply safely, and commit with clear review steps.
- Keep UI responsive and legible for long histories and long messages.

## Tech Stack
- Platform: Wails v2 (Go 1.22+). Native desktop build with web UI.
- Backend (Go): GitService, ChatService, FS/DiffService, SessionStore, Settings, Events.
- Frontend: Vite + Svelte (locked). Implements the chat/graph layout described in `frontend/frontend_description.md`.

## High-Level Architecture
- Wails App (Go) exposes methods/events to the frontend via bindings and runtime.Events.
- Services:
  - GitService: shell out to git for reliability; operations: open repo, status, log (graph), show/diff, blame (later), checkout/branch/merge, apply patch, commit; must support `--allow-empty` commits and standardized commit message formatting with role prefixes and footers.
  - ChatService: provider adapter (ShellCommand first-class; LLMs optional), prompt assembly, streaming tokens, tool-call bridging for git/file ops; ensures every chat turn creates a commit (`USER:`/`AGENT:`) and, upon approval of code changes, creates a follow-up patch commit with `Relates-To` linking. Messages are commits.
  - FSService: read/write files within repo, patch application guardrails, path sanitization.
  - DiffService: generate unified diffs, structured hunks, syntax-highlight metadata for viewer.
  - SessionStore: local persistence for sessions/UI state (expanded threads, filters, selections), lightweight indices/caches, and provider config. No separate Message table — messages are commits. Storage can be SQLite/bbolt or repo-scoped JSON/Git notes.
  - Settings: key management via environment variables only (no keychain/keyring), model selection, safety limits.
  - Events: progress and streaming updates to UI (token stream, git operations).

### Provider: ShellCommand (first-class)
- Config: command template (string), args, timeout, allowlist, working directory = repo root, env vars (safe subset).
- Execution: spawn process, stream stdout to UI as tokens; on exit create `AGENT:` empty commit with full output as body (auto-add prefix if missing). Non-zero exit → commit `Type: message|error` with stderr summary.
- Guardrails: allowlist commands, max output size, timeouts, kill on cancel, sanitize env, never execute outside repo.

## Frontend Surfaces
- Chat View: spine of commits/messages with expandable side histories at merges (container queries + SVG connectors). Each bubble clearly shows role (`USER`/`AGENT`). Patch commits appear attached to or following their originating message bubble.
- Repo Explorer: tree view, file open, quick search.
- Diff Viewer: side-by-side or inline, hunk navigation, apply/rollback per hunk.
- Patch Review: AI-proposed changes → preview → apply to working tree → stage/commit.
- Settings Modal: provider keys, model options, safety toggles.

## Core Flows
1) Open Repo → index minimal metadata → render spine (first-parent) with ability to expand side histories to merge-base.
2) User sends message → create empty commit with `USER:` prefix; agent responds via provider (default: ShellCommand) → stream output and on completion create empty `AGENT:` commit (body = output).
3) If a message leads to code changes: preview diff → on approval create a patch commit using the appropriate role prefix and add `Relates-To: <message-sha>`.
4) Manual file edits: commit with `USER:` prefix, summarizing the change; optionally relate to the last message if applicable.
5) Browse Commits → open diff/files → expand other parent at merges to view side thread.
6) Branch/PR Flow (local): create/switch branch, compare branches, merge with preview.

## Data Model (minimal)
- Commit {sha, parents[], author, date, subject, body, refs[], role(USER|AGENT), type(message|patch|manual-edit), relatesTo?}
- Patch {files[], hunks[]}
- UIState {expandedThreads, filters, selection}  // optional local persistence
- Session {id, repoPath, createdAt, settingsRef} // optional; may also be recorded via commit footer `Session:`
- Settings {provider, model, keyRef, limits}

## Security & Safety
- Never execute arbitrary commands from the model; only whitelisted git/file ops.
- Path sanitization; disallow writes outside repo.
- API keys are sourced from environment variables only; never persist secrets.
- Pre-apply validation for patches; dry-run where possible.

## Testing Strategy
- Unit: GitService against temp repos (including `--allow-empty` message commits and `Relates-To` footers); DiffService golden tests; ChatService adapter mocks.
- Integration: end-to-end flows (open repo → message commit(s) → propose patch → approve → patch commit).
- UI: snapshot tests for chat rendering and graph expand/collapse.

## Performance Notes
- Cache parsed logs; lazy-load diffs on demand; virtualize long lists if needed.
- Stream tokens/events; debounce expensive graph/connector redraws.

## Milestones & Acceptance
- M0: Scaffold Wails app, choose frontend, wire bindings/events; stub services.
  - Done when app builds, opens a window, can select a repo, and prints basic git status.
- M1: GitService + Repo Explorer.
  - Done when repo opens, lists files, shows status, and can view commit log (spine only) with role-prefixed subjects rendered.
- M2: Diff Viewer + File Viewer.
  - Done when a commit/file diff renders with hunk navigation and syntax colors.
- M3: ChatService (ShellCommand stream) + Prompt Bar + Message Commits.
  - Done when user/agent messages create `USER:`/`AGENT:` empty commits; ShellCommand output streams to UI; adapters pluggable.
- M4: Patch Proposal → Preview → Apply → Patch Commit.
  - Done when model-proposed changes can be previewed and, on approval, committed with a role prefix and `Relates-To` link to the originating message commit.
- M5: Graph UI per `frontend/frontend_description.md` (expandable side histories with SVG connectors).
  - Done when merges can expand/collapse to show side threads up to merge-base across breakpoints.
- M6: Persistence & Settings.
  - Done when sessions and settings persist locally; provider keys are read from env and never persisted.
- M7: Packaging & QA.
  - Done when macOS build is packaged; basic test suite passes; crash logs captured.

## Open Decisions
- Frontend: Vite + Svelte (locked).
- Persistence: SQLite vs bbolt (or none). If omitted, use Git notes or repo-scoped JSON for UI state.
- Providers: ShellCommand (default); LLM adapters optional.
- OS coverage: macOS + Linux (Windows out of scope for now).

## OS Coverage: macOS + Linux
- Shell: prefer executing commands directly; when a shell is needed, use `/bin/sh -c` for POSIX portability; avoid bashisms.
- Key storage: environment variables only; no keychain/keyring.
- Packaging: macOS (app bundle); Linux (AppImage/Deb). File watching and paths differ slightly; keep repo operations via git to minimize OS variance.

## Risks/Mitigations
- Large repos: paginate logs and lazy-load diffs; cap history range by default.
- Prompt/tool misuse: strict tool whitelist; confirm before write/commit; clear UI affordances.
- Diff rendering performance: virtualize long diffs; chunked rendering.
