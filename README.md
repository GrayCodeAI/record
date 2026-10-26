# Record

Commit-linked checkpoints for GrayCode agents. Record is a Go library and CLI. A checkpoint records a Git commit, files changed in that commit, a task and session reference, a human-readable summary, and verification evidence. It stores no raw prompt or transcript.

## Try it

```sh
go build -o record ./cmd/record
./record create --repo ../reflect --session <reflect-session-id> --task ISSUE-123 \
  --summary 'Fixed provider preflight' --verification 'go test ./...: passed'
./record list --repo ../reflect
./record show --repo ../reflect --commit HEAD
```

Checkpoints live under the repository's Git common directory in `record/checkpoints/`, so worktrees share them. Files use mode `0600`. `show <id>` also accepts a checkpoint ID. Each checkpoint is immutable; multiple checkpoints can reference the same commit.

## Boundary

Reflect can create checkpoints after a commit and use their IDs as Recall sources. Ridge can display Record's JSON output. Relay has no checkpoint storage role. Record has no compile-time dependency on those products.

The first release is explicit and local: `create` is manual, and checkpoints do not sync to a Git remote. Automatic agent hooks, commit trailers, redacted transcript capture, and cross-machine sync are future work.
