Got it! Here’s a ready-to-paste package you can drop into a Wails/Lovable mock. It’s opinionated but pragmatic: the structures mirror what libgit2/go-git (“gitlib” in many stacks) typically expose, plus a few precomputed helpers (first-parent order, merge-bases, lane hints) that make your chat-style UI easier to wire up.

⸻

🧭 Preamble: Design Intent

Project: Chat-style Git Conversation Interface (Graph-aware, responsive)
Goal: Provide a rich, realistic fake repo graph (branches, merges, octopus merge, cherry-pick, tags, remotes, diffs, co-authors, signed commits) that is:
	•	Close to library models (libgit2/go-git): commits, parents, trees, authors, refs.
	•	UI-ready: includes firstParentOrder, mergeBases, and optional laneHint so you can render the “linear spine + expandable side threads” with minimal glue code.
	•	Human-narrative friendly: long commit messages, fixup/squash chains, reverts, release flow, and realistic timestamps.

Use the TypeScript types if you’re in React/Lovable; or just treat the JSON as your mock API response.

⸻

🧩 TypeScript Model (drop-in)

// — Core commit/refs shape, close to libgit2/go-git —
export type Sha = string;

export interface Identity {
  name: string;
  email: string;
  when: string; // ISO 8601
  tzOffsetMinutes?: number;
}

export interface FileChange {
  path: string;
  oldPath?: string;         // for renames
  status: "added"|"modified"|"deleted"|"renamed"|"copied"|"typechange";
  additions: number;
  deletions: number;
  binary?: boolean;
}

export interface DiffStat {
  filesChanged: number;
  insertions: number;
  deletions: number;
}

export interface Commit {
  id: Sha;
  tree: Sha;
  parents: Sha[];           // 0=root; 1=normal; 2+=merge
  author: Identity;
  committer: Identity;
  message: string;          // include full body (co-authored-by, signed-off-by)
  summary: string;          // first line
  notes?: string[];         // git-notes (optional)
  gpg?: { verified: boolean; keyId?: string; who?: Identity };
  stats?: DiffStat;         // aggregate
  changes?: FileChange[];   // per-file
  extra?: {
    // helpers for UI
    firstParentIndex?: number;  // index in first-parent linearization
    laneHint?: number;          // suggested side lane for side-thread (optional)
    kind?: "merge"|"octopus"|"revert"|"fixup"|"squash"|"cherry-pick"|"normal";
  };
}

export interface BranchRef {
  name: string;             // e.g., "main", "feature/user-auth"
  fullName: string;         // e.g., "refs/heads/main"
  head: Sha;
  upstream?: { remote: string; name: string }; // remote tracking
  isHEAD?: boolean;         // currently checked out
}

export interface TagRef {
  name: string;             // "v1.2.0"
  fullName: string;         // "refs/tags/v1.2.0"
  target: Sha;              // object id (commit)
  annotated?: boolean;
  tagger?: Identity;
  message?: string;
}

export interface RemoteRef {
  name: string;             // "origin"
  url: string;              // "git@github.com:org/repo.git"
  fetch?: string;
  push?: string;
}

export interface MergeBase {
  a: Sha;
  b: Sha;
  base: Sha;
}

export interface RepoGraph {
  repo: {
    id: string;
    name: string;
    defaultBranch: string;  // "main"
    description?: string;
  };
  remotes: RemoteRef[];
  branches: BranchRef[];
  tags: TagRef[];
  commits: Record<Sha, Commit>;
  // Helpful precomputations for UI:
  firstParentOrder: Sha[];  // from oldest -> newest along default branch
  mergeBases: MergeBase[];  // to draw faint confluence connectors
  // Optional: handy groupings for your chat lanes/threads
  threads?: {
    // tip -> list of SHAs in that side history (oldest->newest)
    [branchName: string]: Sha[];
  };
}


⸻

📦 Mock Data (realistic, rich, and UI-ready)

Timestamps are in Europe/Berlin (UTC+02:00 right now).
Today is 2025-10-12; history spans ~2 months for realism.

{
  "repo": {
    "id": "mock-graph-001",
    "name": "chat-git-ui",
    "defaultBranch": "main",
    "description": "Demo repo for chat-style Git conversation UI with rich graph cases."
  },
  "remotes": [
    {
      "name": "origin",
      "url": "git@github.com:example/chat-git-ui.git",
      "fetch": "+refs/heads/*:refs/remotes/origin/*"
    }
  ],
  "branches": [
    { "name": "main", "fullName": "refs/heads/main", "head": "b1d3b1d3", "upstream": { "remote": "origin", "name": "main" }, "isHEAD": true },
    { "name": "feature/user-auth", "fullName": "refs/heads/feature/user-auth", "head": "9af0c9e2", "upstream": { "remote": "origin", "name": "feature/user-auth" } },
    { "name": "fix/bug-123", "fullName": "refs/heads/fix/bug-123", "head": "6e78c4f1" },
    { "name": "chore/ci", "fullName": "refs/heads/chore/ci", "head": "7f55aa00" },
    { "name": "release/v1.2", "fullName": "refs/heads/release/v1.2", "head": "f0f0a1a1" },
    { "name": "hotfix/urgent-500", "fullName": "refs/heads/hotfix/urgent-500", "head": "cafe5000" },
    { "name": "experiment/imgui", "fullName": "refs/heads/experiment/imgui", "head": "e11a1a11" },
    { "name": "docs/readme-polish", "fullName": "refs/heads/docs/readme-polish", "head": "da7a5eed" }
  ],
  "tags": [
    {
      "name": "v1.1.0",
      "fullName": "refs/tags/v1.1.0",
      "target": "2aa2aa22",
      "annotated": true,
      "tagger": { "name": "CI Bot", "email": "ci@bot.local", "when": "2025-09-05T10:12:03+02:00" },
      "message": "Release v1.1.0\n\nChangelog:\n- Basic chat layout\n- Initial graph overlay"
    },
    {
      "name": "v1.2.0-rc1",
      "fullName": "refs/tags/v1.2.0-rc1",
      "target": "f0f0a1a1",
      "annotated": true,
      "tagger": { "name": "Release Eng", "email": "releng@org.local", "when": "2025-10-08T17:40:11+02:00" },
      "message": "v1.2.0-rc1: responsive side lanes; expandable side histories."
    }
  ],
  "commits": {
    "00000001": {
      "id": "00000001",
      "tree": "t0001",
      "parents": [],
      "author": { "name": "Alice", "email": "alice@example.com", "when": "2025-08-20T09:02:10+02:00" },
      "committer": { "name": "Alice", "email": "alice@example.com", "when": "2025-08-20T09:02:10+02:00" },
      "summary": "init: create Wails/React project skeleton",
      "message": "init: create Wails/React project skeleton\n\nIncludes basic routes and placeholder chat view.\n",
      "stats": { "filesChanged": 7, "insertions": 713, "deletions": 0 },
      "changes": [
        { "path": "wails.json", "status": "added", "additions": 210, "deletions": 0 },
        { "path": "frontend/src/main.tsx", "status": "added", "additions": 102, "deletions": 0 },
        { "path": "go.mod", "status": "added", "additions": 18, "deletions": 0 }
      ],
      "extra": { "kind": "normal", "firstParentIndex": 0 }
    },
    "11111111": {
      "id": "11111111",
      "tree": "t0002",
      "parents": ["00000001"],
      "author": { "name": "Bob", "email": "bob@example.com", "when": "2025-08-21T14:14:05+02:00" },
      "committer": { "name": "Bob", "email": "bob@example.com", "when": "2025-08-21T14:20:19+02:00" },
      "summary": "feat(graph): first-parent spine rendering",
      "message": "feat(graph): first-parent spine rendering\n\nDraw simple vertical spine; stub overlay.\nCo-authored-by: Carol <carol@example.com>\n",
      "stats": { "filesChanged": 5, "insertions": 188, "deletions": 12 },
      "extra": { "kind": "normal", "firstParentIndex": 1 }
    },
    "22222222": {
      "id": "22222222",
      "tree": "t0003",
      "parents": ["11111111"],
      "author": { "name": "Alice", "email": "alice@example.com", "when": "2025-08-22T09:47:33+02:00" },
      "committer": { "name": "Alice", "email": "alice@example.com", "when": "2025-08-22T09:50:02+02:00" },
      "summary": "chore(ci): add GitHub Actions build",
      "message": "chore(ci): add GitHub Actions build\n\nSigned-off-by: Alice <alice@example.com>\n",
      "stats": { "filesChanged": 2, "insertions": 39, "deletions": 0 },
      "extra": { "kind": "normal", "firstParentIndex": 2 }
    },
    "2aa2aa22": {
      "id": "2aa2aa22",
      "tree": "t0004",
      "parents": ["22222222"],
      "author": { "name": "CI Bot", "email": "ci@bot.local", "when": "2025-09-05T10:10:00+02:00" },
      "committer": { "name": "CI Bot", "email": "ci@bot.local", "when": "2025-09-05T10:11:59+02:00" },
      "summary": "release: v1.1.0",
      "message": "release: v1.1.0\n\nTag: v1.1.0",
      "stats": { "filesChanged": 1, "insertions": 1, "deletions": 0 },
      "extra": { "kind": "normal", "firstParentIndex": 3 }
    },

    "33333333": {
      "id": "33333333",
      "tree": "t0100",
      "parents": ["2aa2aa22"],
      "author": { "name": "Carol", "email": "carol@example.com", "when": "2025-09-10T13:05:41+02:00" },
      "committer": { "name": "Carol", "email": "carol@example.com", "when": "2025-09-10T13:11:00+02:00" },
      "summary": "docs: README polish and screenshots",
      "message": "docs: README polish and screenshots\n",
      "stats": { "filesChanged": 3, "insertions": 97, "deletions": 10 },
      "extra": { "kind": "normal", "firstParentIndex": 4 }
    },
    "44444444": {
      "id": "44444444",
      "tree": "t0200",
      "parents": ["33333333"],
      "author": { "name": "Dave", "email": "dave@example.com", "when": "2025-09-12T16:42:12+02:00" },
      "committer": { "name": "Dave", "email": "dave@example.com", "when": "2025-09-12T16:45:33+02:00" },
      "summary": "feat(ui): container queries for side lanes",
      "message": "feat(ui): container queries for side lanes\n\nEnables medium/wide layouts.",
      "stats": { "filesChanged": 6, "insertions": 240, "deletions": 18 },
      "extra": { "kind": "normal", "firstParentIndex": 5 }
    },

    "a1a1a1a1": {
      "id": "a1a1a1a1",
      "tree": "t1000",
      "parents": ["33333333"],
      "author": { "name": "Bob", "email": "bob@example.com", "when": "2025-09-11T10:00:00+02:00" },
      "committer": { "name": "Bob", "email": "bob@example.com", "when": "2025-09-11T10:05:00+02:00" },
      "summary": "feat(auth): login form scaffolding",
      "message": "feat(auth): login form scaffolding\n",
      "stats": { "filesChanged": 4, "insertions": 120, "deletions": 0 },
      "extra": { "kind": "normal" }
    },
    "a2a2a2a2": {
      "id": "a2a2a2a2",
      "tree": "t1001",
      "parents": ["a1a1a1a1"],
      "author": { "name": "Bob", "email": "bob@example.com", "when": "2025-09-11T12:20:00+02:00" },
      "committer": { "name": "Bob", "email": "bob@example.com", "when": "2025-09-11T12:25:00+02:00" },
      "summary": "fixup! feat(auth): login form scaffolding",
      "message": "fixup! feat(auth): login form scaffolding\n",
      "extra": { "kind": "fixup" }
    },
    "a3a3a3a3": {
      "id": "a3a3a3a3",
      "tree": "t1002",
      "parents": ["a2a2a2a2"],
      "author": { "name": "Bob", "email": "bob@example.com", "when": "2025-09-11T18:04:00+02:00" },
      "committer": { "name": "Bob", "email": "bob@example.com", "when": "2025-09-11T18:12:00+02:00" },
      "summary": "feat(auth): OAuth device code flow",
      "message": "feat(auth): OAuth device code flow\n\nCo-authored-by: Eve <eve@example.com>\n",
      "extra": { "kind": "normal" }
    },
    "9af0c9e2": {
      "id": "9af0c9e2",
      "tree": "t1099",
      "parents": ["a3a3a3a3", "44444444"],
      "author": { "name": "Bob", "email": "bob@example.com", "when": "2025-09-13T09:20:00+02:00" },
      "committer": { "name": "Bob", "email": "bob@example.com", "when": "2025-09-13T09:22:30+02:00" },
      "summary": "merge: feature/user-auth -> main",
      "message": "merge: feature/user-auth -> main\n\nResolves #42.\n",
      "stats": { "filesChanged": 8, "insertions": 310, "deletions": 25 },
      "extra": { "kind": "merge", "firstParentIndex": 6 }
    },

    "55555555": {
      "id": "55555555",
      "tree": "t0300",
      "parents": ["9af0c9e2"],
      "author": { "name": "Eve", "email": "eve@example.com", "when": "2025-09-15T11:11:11+02:00" },
      "committer": { "name": "Eve", "email": "eve@example.com", "when": "2025-09-15T11:11:40+02:00" },
      "summary": "refactor: extract SvgOverlay component",
      "message": "refactor: extract SvgOverlay component\n\nImproves connector rendering.\n",
      "stats": { "filesChanged": 5, "insertions": 172, "deletions": 60 },
      "extra": { "kind": "normal", "firstParentIndex": 7 }
    },

    "b00b5eed": {
      "id": "b00b5eed",
      "tree": "t0400",
      "parents": ["55555555"],
      "author": { "name": "Dave", "email": "dave@example.com", "when": "2025-09-16T08:25:00+02:00" },
      "committer": { "name": "Dave", "email": "dave@example.com", "when": "2025-09-16T08:29:00+02:00" },
      "summary": "revert: \"feat(ui): container queries for side lanes\"",
      "message": "This reverts commit 44444444.\n\nReason: regression on Safari.",
      "extra": { "kind": "revert", "firstParentIndex": 8 }
    },

    "66cc66cc": {
      "id": "66cc66cc",
      "tree": "t0500",
      "parents": ["b00b5eed"],
      "author": { "name": "Alice", "email": "alice@example.com", "when": "2025-09-17T17:00:00+02:00" },
      "committer": { "name": "Alice", "email": "alice@example.com", "when": "2025-09-17T17:05:00+02:00" },
      "summary": "feat(layout): safe container queries w/ fallbacks",
      "message": "feat(layout): safe container queries w/ fallbacks\n\nRe-implements 44444444 safely.",
      "extra": { "kind": "normal", "firstParentIndex": 9 }
    },

    "7f55aa00": {
      "id": "7f55aa00",
      "tree": "t2000",
      "parents": ["2aa2aa22"],
      "author": { "name": "CI Bot", "email": "ci@bot.local", "when": "2025-09-06T06:06:06+02:00" },
      "committer": { "name": "CI Bot", "email": "ci@bot.local", "when": "2025-09-06T06:06:06+02:00" },
      "summary": "ci: split jobs (lint/test/build)",
      "message": "ci: split jobs (lint/test/build)\n",
      "extra": { "kind": "normal" }
    },
    "7f55aa11": {
      "id": "7f55aa11",
      "tree": "t2001",
      "parents": ["7f55aa00"],
      "author": { "name": "CI Bot", "email": "ci@bot.local", "when": "2025-09-07T07:10:00+02:00" },
      "committer": { "name": "CI Bot", "email": "ci@bot.local", "when": "2025-09-07T07:11:00+02:00" },
      "summary": "ci: cache pnpm store",
      "message": "ci: cache pnpm store\n",
      "extra": { "kind": "normal" }
    },
    "7f55aa22": {
      "id": "7f55aa22",
      "tree": "t2002",
      "parents": ["7f55aa11", "33333333", "22222222"],
      "author": { "name": "CI Bot", "email": "ci@bot.local", "when": "2025-09-08T08:08:08+02:00" },
      "committer": { "name": "CI Bot", "email": "ci@bot.local", "when": "2025-09-08T08:08:08+02:00" },
      "summary": "merge (octopus): bring docs+ci+core into chore/ci",
      "message": "merge (octopus): combine maintenance lines before release.\n",
      "extra": { "kind": "octopus" }
    },

    "6e78c4f1": {
      "id": "6e78c4f1",
      "tree": "t3000",
      "parents": ["55555555"],
      "author": { "name": "Eve", "email": "eve@example.com", "when": "2025-09-15T14:33:00+02:00" },
      "committer": { "name": "Eve", "email": "eve@example.com", "when": "2025-09-15T14:38:00+02:00" },
      "summary": "fix: bug-123 guard null lane placements",
      "message": "fix: bug-123 guard null lane placements\n\nCloses #123.\n",
      "extra": { "kind": "normal" }
    },

    "f0f0a1a1": {
      "id": "f0f0a1a1",
      "tree": "t4000",
      "parents": ["66cc66cc"],
      "author": { "name": "Release Eng", "email": "releng@org.local", "when": "2025-10-08T17:37:00+02:00" },
      "committer": { "name": "Release Eng", "email": "releng@org.local", "when": "2025-10-08T17:39:00+02:00" },
      "summary": "chore(release): prepare v1.2 branch",
      "message": "chore(release): prepare v1.2 branch\n\nBackports minimal fixes.",
      "extra": { "kind": "normal" }
    },

    "cafe5000": {
      "id": "cafe5000",
      "tree": "t5000",
      "parents": ["66cc66cc"],
      "author": { "name": "Ops", "email": "ops@example.com", "when": "2025-10-09T08:00:00+02:00" },
      "committer": { "name": "Ops", "email": "ops@example.com", "when": "2025-10-09T08:01:00+02:00" },
      "summary": "hotfix: urgent 500 on /auth callback (cherry-pick of a3a3a3a3)",
      "message": "cherry-pick: a3a3a3a3\n\nMinimal patch only.",
      "extra": { "kind": "cherry-pick" }
    },

    "e11a1a11": {
      "id": "e11a1a11",
      "tree": "t6000",
      "parents": ["33333333"],
      "author": { "name": "Researcher", "email": "r@example.com", "when": "2025-09-10T19:20:00+02:00" },
      "committer": { "name": "Researcher", "email": "r@example.com", "when": "2025-09-10T19:21:00+02:00" },
      "summary": "experiment: Dear ImGui proto",
      "message": "experiment: Dear ImGui proto\n\nStreaming panel, docking playground.",
      "extra": { "kind": "normal" }
    },

    "da7a5eed": {
      "id": "da7a5eed",
      "tree": "t7000",
      "parents": ["33333333"],
      "author": { "name": "Carol", "email": "carol@example.com", "when": "2025-09-10T15:32:00+02:00" },
      "committer": { "name": "Carol", "email": "carol@example.com", "when": "2025-09-10T15:33:00+02:00" },
      "summary": "docs: README grammar & tone pass",
      "message": "docs: README grammar & tone pass\n",
      "extra": { "kind": "normal" }
    },

    "b1d3b1d3": {
      "id": "b1d3b1d3",
      "tree": "t9000",
      "parents": ["66cc66cc"],
      "author": { "name": "Alice", "email": "alice@example.com", "when": "2025-10-10T12:12:12+02:00" },
      "committer": { "name": "Alice", "email": "alice@example.com", "when": "2025-10-10T12:14:00+02:00" },
      "summary": "feat(a11y): keyboard shortcuts (E/[,],Esc)",
      "message": "feat(a11y): keyboard shortcuts (E/[,],Esc)\n\nARIA labels for side histories.",
      "extra": { "kind": "normal", "firstParentIndex": 10 }
    }
  },

  "firstParentOrder": [
    "00000001",
    "11111111",
    "22222222",
    "2aa2aa22",
    "33333333",
    "44444444",
    "9af0c9e2",
    "55555555",
    "b00b5eed",
    "66cc66cc",
    "b1d3b1d3"
  ],

  "mergeBases": [
    { "a": "9af0c9e2", "b": "a3a3a3a3", "base": "33333333" },
    { "a": "7f55aa22", "b": "33333333", "base": "33333333" },
    { "a": "7f55aa22", "b": "22222222", "base": "22222222" },
    { "a": "b00b5eed", "b": "44444444", "base": "33333333" }
  ],

  "threads": {
    "feature/user-auth": ["a1a1a1a1", "a2a2a2a2", "a3a3a3a3", "9af0c9e2"],
    "chore/ci": ["7f55aa00", "7f55aa11", "7f55aa22"],
    "fix/bug-123": ["6e78c4f1"],
    "release/v1.2": ["f0f0a1a1"],
    "hotfix/urgent-500": ["cafe5000"],
    "experiment/imgui": ["e11a1a11"],
    "docs/readme-polish": ["da7a5eed"]
  }
}


⸻

🔧 Glue Notes (Wails / go-git / libgit2)
	•	Mapping:
	•	Commit.parents → libgit2 git_commit_parentcount() / go-git Commit.Parents.
	•	author/committer → git_signature.
	•	tags/branches/remotes → git_reference (e.g., refs/heads/*, refs/tags/*).
	•	Octopus merge is represented by parents.length > 2 (7f55aa22).
	•	Cherry-pick (cafe5000) uses a message hint (you can add a cherryPickOf: Sha if you prefer).
	•	Revert (b00b5eed) keeps the standard “This reverts …” message for easy detection.
	•	First-parent linearization is precomputed as firstParentOrder to drive the chat “spine”.
	•	Merge bases help you draw faint “expanded side history up to confluence” connectors.

⸻

🧪 Mini Slice (for unit tests)

{
  "branches": [
    { "name": "main", "fullName": "refs/heads/main", "head": "C" },
    { "name": "feature/x", "fullName": "refs/heads/feature/x", "head": "B2" }
  ],
  "commits": {
    "A": { "id": "A", "tree": "tA", "parents": [], "author": { "name": "a", "email": "a@x", "when": "2025-01-01T00:00:00+01:00" }, "committer": { "name": "a", "email": "a@x", "when": "2025-01-01T00:00:00+01:00" }, "summary": "init", "message": "init" },
    "B1": { "id": "B1", "tree": "tB1", "parents": ["A"], "author": { "name": "b", "email": "b@x", "when": "2025-01-02T00:00:00+01:00" }, "committer": { "name": "b", "email": "b@x", "when": "2025-01-02T00:00:00+01:00" }, "summary": "main work 1", "message": "main work 1" },
    "B2": { "id": "B2", "tree": "tB2", "parents": ["A"], "author": { "name": "b", "email": "b@x", "when": "2025-01-02T01:00:00+01:00" }, "committer": { "name": "b", "email": "b@x", "when": "2025-01-02T01:00:00+01:00" }, "summary": "feature work 1", "message": "feature work 1" },
    "C":  { "id": "C",  "tree": "tC",  "parents": ["B1","B2"], "author": { "name": "m", "email": "m@x", "when": "2025-01-03T00:00:00+01:00" }, "committer": { "name": "m", "email": "m@x", "when": "2025-01-03T00:00:00+01:00" }, "summary": "merge feature/x", "message": "merge feature/x", "extra": { "kind": "merge" } }
  },
  "firstParentOrder": ["A","B1","C"],
  "mergeBases": [{ "a": "C", "b": "B2", "base": "A" }]
}


⸻

✅ How to Use It Fast
	•	In Wails (Go): serve this JSON from a mock handler; your frontend calls it as if from a real git backend.
	•	In Lovable/React: copy the TypeScript types + JSON into state; render your chat spine with firstParentOrder, and mount side threads from threads[branchName] or by walking parents[] until mergeBases[].base.

If you want this reshaped to go-structs for a Wails backend or a Redux slice for the frontend, say the word and I’ll output that variant too.