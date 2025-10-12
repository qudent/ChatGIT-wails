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

See `types.ts` for the complete TypeScript interface definitions.


⸻

📦 Mock Data (realistic, rich, and UI-ready)

See `mock-data.json` for the complete mock data JSON.

🧪 Mini Slice (for unit tests)

See `mini-test.json` for the mini test slice JSON.

✅ How to Use It Fast
	•	In Wails (Go): serve this JSON from a mock handler; your frontend calls it as if from a real git backend.
	•	In Lovable/React: copy the TypeScript types + JSON into state; render your chat spine with firstParentOrder, and mount side threads from threads[branchName] or by walking parents[] until mergeBases[].base.

If you want this reshaped to go-structs for a Wails backend or a Redux slice for the frontend, say the word and I’ll output that variant too.