10. [ ] Create internal packages: internal/services/{git,fs,diff,chat,session,settings} and internal/models


20. [ ] Add process exec helper with context timeout and a git command whitelist; capture stdout/stderr


30. [ ] Bind service facades to Wails (options.Bind): GitService, FSService, DiffService, ChatService, SettingsService


40. [ ] Implement Config and Logger (levels) and inject into services


50. [ ] GitService: SetRepo(path) with validation; verify .git present and store repo root


60. [ ] GitService: GitVersion() and Health() checks


70. [ ] GitService: Status() (branch, ahead/behind, staged/unstaged counts)


80. [ ] Models: Commit struct per PLAN (sha, parents, author, subject, body, refs, role, type, relatesTo)


90. [ ] GitService: LogSpine(limit, skip) first-parent; parse role/type from subject


100. [ ] GitService: LogWithGraph(options) including parents[] and refs[]; support ranges and merges


110. [ ] GitService: ShowFile(commitSha, path) and ListTree(commitSha, path)


120. [ ] GitService: CreateEmptyCommit(role, subject, body, footers) with --allow-empty


130. [ ] GitService: ApplyPatch(unifiedDiff, allowCreate) dry-run then apply; stage and commit


140. [ ] GitService: AddPatchCommit(role, relatesToSha, summary, body) with Relates-To footer


150. [ ] GitService: Branch list/create/checkout; MergePreview(base, head) using --no-commit


160. [ ] GitService: ComputeMergeBase(a, b) for side-thread expansion


170. [ ] DiffService: DiffCommit(commitSha) unified diff; parse to hunks metadata for UI


180. [ ] DiffService: DiffRange(base, head) and DiffWorkingTree()


190. [ ] FSService: SafeRead(path) with repo-root enforcement and path sanitization


200. [ ] FSService: SafeWrite(path, data) with validations; optional dry-run


210. [ ] ChatService: Provider adapter interface; stub OpenAI adapter; config-driven selection


220. [ ] ChatService: StartMessage(role, text) → stream tokens via runtime.Events


230. [ ] ChatService: OnUserMessage → create USER empty commit; OnAgentFinal → create AGENT empty commit


240. [ ] ChatService: ProposePatch(flow) → return diff for UI review; upon approval call AddPatchCommit


250. [ ] SettingsService: Get/Set provider, model, limits; key retrieval from env/OS keychain (no plaintext)


260. [ ] SessionStore: Persist UIState and session metadata (JSON or bbolt) scoped to repo


270. [ ] Events: Define and emit progress events for git ops and patch apply


280. [ ] Security: Enforce command whitelist; sanitize paths; block writes outside repo


290. [ ] Performance: Cache parsed logs; add limits/pagination to log queries


300. [ ] Wails wiring: file dialogs to select repo; call SetRepo; display basic status


310. [ ] Test helpers to init temp repos and common commit/branch scenarios


320. [ ] Unit: GitService (allow-empty, role parsing, Relates-To), DiffService golden tests, FSService sanitization


330. [ ] Unit: ChatService with mocked provider and event stream


340. [ ] Integration: Open repo → message commits → propose patch → approve → patch commit


350. [ ] CI/local scripts: lint, go vet, go test


360. [ ] Packaging: macOS bundle via wails build; verify app icon and About info
