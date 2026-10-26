package checkpoint

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestCommitLinkedCheckpointAcrossWorktrees(t *testing.T) {
	repo := t.TempDir()
	gitTest(t, repo, "init", "-b", "main")
	gitTest(t, repo, "config", "user.name", "Test")
	gitTest(t, repo, "config", "user.email", "test@example.com")
	if err := os.WriteFile(filepath.Join(repo, "file.txt"), []byte("hello"), 0600); err != nil {
		t.Fatal(err)
	}
	gitTest(t, repo, "add", "file.txt")
	gitTest(t, repo, "commit", "-m", "Initial commit")
	s := Store{Repo: repo}
	r, err := s.Create("session-1", "TASK-1", "Created file", []string{"go test ./...: passed"})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Files) != 1 || r.Files[0] != "file.txt" || r.Commit == "" {
		t.Fatalf("missing commit evidence: %#v", r)
	}
	_, dir, err := s.gitLocations()
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(filepath.Join(dir, r.ID+".json"))
	if err != nil || info.Mode().Perm() != 0600 {
		t.Fatalf("checkpoint permissions: %v, %v", info, err)
	}
	worktree := filepath.Join(t.TempDir(), "other")
	gitTest(t, repo, "worktree", "add", "-b", "other", worktree)
	other := Store{Repo: worktree}
	got, err := other.List("HEAD")
	if err != nil || len(got) != 1 || got[0].ID != r.ID {
		t.Fatalf("checkpoint not shared across worktrees: %#v, %v", got, err)
	}
}

func TestCheckpointRequiresCommitAndSummary(t *testing.T) {
	repo := t.TempDir()
	gitTest(t, repo, "init", "-b", "main")
	s := Store{Repo: repo}
	if _, err := s.Create("", "", "", nil); err == nil {
		t.Fatal("accepted missing summary")
	}
	if _, err := s.Create("", "", "summary", nil); err == nil {
		t.Fatal("accepted repository without a commit")
	}
	if _, err := s.Get("../../secret"); err == nil {
		t.Fatal("accepted traversal ID")
	}
}

func gitTest(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
}
