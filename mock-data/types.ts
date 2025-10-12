

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


