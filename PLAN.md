# Plan: Git-based ChatGPT (Wails + Go)

## Goals
- Desktop app that treats a Git repo like a chat: main branch as the narrative spine, merges as expandable side threads, commits/messages as bubbles with long text.
- Let the AI propose patches, preview diffs, apply safely, and commit with clear review steps.
- Keep UI responsive and legible for long histories and long messages.

## Tech Stack
- Platform: Wails v2 (Go 1.22+). Native desktop build with web UI.
- Backend (Go): GitService, ChatService, FS/DiffService, SessionStore, Settings, Events.
- Frontend: pick one at Milestone 0 (default: React; alternatives: Svelte/Vue). Implements the chat/graph layout described in `frontend/frontend_description.md`.

## High-Level Architecture
- Wails App (Go) exposes methods/events to the frontend via bindings and runtime.Events.
- Services:
  - GitService: shell out to git for reliability; operations: open repo, status, log (graph), show/diff, blame (later), checkout/branch/merge, apply patch, commit.
  - ChatService: provider adapter (OpenAI/others), prompt assembly, streaming tokens, tool-call bridging for git/file ops.
  - FSService: read/write files within repo, patch application guardrails, path sanitization.
  - DiffService: generate unified diffs, structured hunks, syntax-highlight metadata for viewer.
  - SessionStore: local persistence for sessions, messages, provider config (SQLite or BoltDB; decide at M0).
  - Settings: key management (prefer OS keychain/env), model selection, safety limits.
  - Events: progress and streaming updates to UI (token stream, git operations).

## Frontend Surfaces
- Chat View: spine of commits/messages with expandable side histories at merges (container queries + SVG connectors).
- Repo Explorer: tree view, file open, quick search.
- Diff Viewer: side-by-side or inline, hunk navigation, apply/rollback per hunk.
- Patch Review: AI-proposed changes → preview → apply to working tree → stage/commit.
- Settings Modal: provider keys, model options, safety toggles.

## Core Flows
1) Open Repo → index minimal metadata → render spine (first-parent) with ability to expand side histories to merge-base.
2) Chat → model proposes patch(s) → preview diff → apply (guarded) → stage/commit on confirm → message recorded as commit and chat item.
3) Browse Commits → open diff/files → expand other parent at merges to view side thread.
4) Branch/PR Flow (local): create/switch branch, compare branches, merge with preview.

## Data Model (minimal)
- Session {id, repoPath, createdAt, settingsRef}
- Message {id, sessionId, role, content, ts, relatedCommit?}
- Commit {sha, parents[], author, date, message, refs[]}
- Patch {files[], hunks[]}
- Settings {provider, model, keyRef, limits}

## Security & Safety
- Never execute arbitrary commands from the model; only whitelisted git/file ops.
- Path sanitization; disallow writes outside repo.
- Key storage via OS keychain or env vars; do not persist secrets in plaintext.
- Pre-apply validation for patches; dry-run where possible.

## Testing Strategy
- Unit: GitService against temp repos; DiffService golden tests; ChatService adapter mocks.
- Integration: end-to-end flows (open repo → propose patch → apply → commit).
- UI: snapshot tests for chat rendering and graph expand/collapse.

## Performance Notes
- Cache parsed logs; lazy-load diffs on demand; virtualize long lists if needed.
- Stream tokens/events; debounce expensive graph/connector redraws.

## Milestones & Acceptance
- M0: Scaffold Wails app, choose frontend, wire bindings/events; stub services.
  - Done when app builds, opens a window, can select a repo, and prints basic git status.
- M1: GitService + Repo Explorer.
  - Done when repo opens, lists files, shows status, and can view commit log (spine only).
- M2: Diff Viewer + File Viewer.
  - Done when a commit/file diff renders with hunk navigation and syntax colors.
- M3: ChatService (stream) + Prompt Bar.
  - Done when tokens stream to UI; mock provider works; adapters pluggable.
- M4: Patch Proposal → Preview → Apply → Commit.
  - Done when model-proposed changes can be previewed, selectively applied, and committed.
- M5: Graph UI per `frontend/frontend_description.md` (expandable side histories with SVG connectors).
  - Done when merges can expand/collapse to show side threads up to merge-base across breakpoints.
- M6: Persistence & Settings.
  - Done when sessions and settings persist locally; provider keys managed securely.
- M7: Packaging & QA.
  - Done when macOS build is packaged; basic test suite passes; crash logs captured.

## Open Decisions
- Frontend framework (default React unless specified).
- Local DB: SQLite vs BoltDB.
- Provider(s) to support first: OpenAI, others.
- OS coverage beyond macOS (Windows/Linux) and related git/path nuances.

## Risks/Mitigations
- Large repos: paginate logs and lazy-load diffs; cap history range by default.
- Prompt/tool misuse: strict tool whitelist; confirm before write/commit; clear UI affordances.
- Diff rendering performance: virtualize long diffs; chunked rendering.
