# Record

Record owns immutable checkpoints linked to Git commits. It records summaries and evidence metadata; it does not store model credentials, make model calls, or maintain long-term memory.

Use a feature branch for changes after the initial repository commit. Keep the checkpoint schema versioned and preserve immutable IDs. Run `go test ./...` before committing.
