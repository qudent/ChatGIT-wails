# TRICKY AREAS & SIMPLIFIED APPROACHES:
# - Command injection: Use go-git library instead of shell git commands (eliminates whitelist need)
# - Path traversal: Use filepath.Abs() + strings.HasPrefix() checks, disallow symlinks
# - Legacy repo compatibility: Make role/type parsing optional with graceful fallbacks
# - Event complexity: Separate channels for git ops vs chat tokens, use simple structs
# - Relates-To integrity: Basic existence validation, graceful failure on missing commits
# - Keychain complexity: Start with env vars only, store encrypted settings in repo
# - Patch atomicity: Use gitstash for rollback, simple file diff validation

10. [ ] Create internal packages: internal/services/{git,fs,diff,chat,session,settings} and internal/models


20. [ ] Use go-git library for all git operations (no shell commands, no injection risk)


30. [ ] Bind service facades to Wails (options.Bind): GitService, FSService, DiffService, ChatService, SettingsService


40. [ ] Implement Config and Logger (levels) and inject into services


50. [ ] GitService: SetRepo(path) using go-git.Open; verify .git present and store repo root


60. [ ] GitService: GitVersion() via go-git; Health() with simple repo validation


70. [ ] GitService: Status() (branch, ahead/behind, staged/unstaged counts) via go-git


80. [ ] Models: Commit struct per PLAN (sha, parents, author, subject, body, refs, role, type, relatesTo)


90. [ ] GitService: LogSpine(limit, skip) first-parent; optional role/type parsing with fallbacks


100. [ ] GitService: LogWithGraph(options) using go-git; support ranges and basic merges


110. [ ] GitService: ShowFile(commitSha, path) and ListTree(commitSha, path) via go-git


120. [ ] GitService: CreateEmptyCommit(role, subject, body, footers) with go-git plumbing


130. [ ] GitService: ApplyPatch(unifiedDiff, allowCreate) using git stash for safety; simple validation


140. [ ] GitService: AddPatchCommit(role, relatesToSha, summary, body) with basic existence check only


150. [ ] GitService: Basic branch list/create/checkout; skip advanced merge preview for now


160. [ ] GitService: ComputeMergeBase(a, b) for side-thread expansion via go-git


170. [ ] DiffService: DiffCommit(commitSha) unified diff via go-git; simple hunk parsing


180. [ ] DiffService: DiffRange(base, head) and DiffWorkingTree() via go-git


190. [ ] FSService: SafeRead(path) with filepath.Abs validation; block symlinks and relative paths


200. [ ] FSService: SafeWrite(path, data) with root boundary checks; allow only file extensions in allowlist


210. [ ] ChatService: Simple provider interface; OpenAI adapter; config-driven selection


220. [ ] ChatService: StartMessage(role, text) → stream tokens via separate channel from git events


230. [ ] ChatService: OnUserMessage → create USER commit; OnAgentFinal → create AGENT commit


240. [ ] ChatService: ProposePatch(flow) → return diff for UI; approval calls AddPatchCommit


250. [ ] SettingsService: Get/Set provider, model, limits; env vars only (no keychain complexity)


260. [ ] SessionStore: Persist UIState as JSON files in repo/.chatgit/ directory


270. [ ] Events: Simple progress structs for git ops; separate channel for chat tokens


280. [ ] Security: go-git eliminates command injection; file allowlist + boundary checks for paths


290. [ ] Performance: Basic in-memory commit cache; pagination for log queries only


300. [ ] Wails wiring: file dialogs to select repo; call SetRepo; display basic status


310. [ ] Test helpers to init temp repos and common commit scenarios


320. [ ] Unit: GitService (go-git operations), DiffService tests, FSService boundary checks


330. [ ] Unit: ChatService with mocked provider and separate event channels


340. [ ] Integration: Open repo → create commits → apply patch workflow end-to-end


350. [ ] CI/local scripts: lint, go vet, go test


360. [ ] Packaging: macOS bundle via wails build; verify app icon and About info
