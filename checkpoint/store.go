package checkpoint

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Record is immutable evidence associated with a committed code change.
type Record struct {
	Version      int       `json:"version"`
	ID           string    `json:"id"`
	Repository   string    `json:"repository"`
	Commit       string    `json:"commit"`
	SessionID    string    `json:"session_id,omitempty"`
	TaskID       string    `json:"task_id,omitempty"`
	Summary      string    `json:"summary"`
	Verification []string  `json:"verification,omitempty"`
	Files        []string  `json:"files"`
	CreatedAt    time.Time `json:"created_at"`
}

type Store struct{ Repo string }

func (s Store) Create(sessionID, taskID, summary string, verification []string) (Record, error) {
	summary = strings.TrimSpace(summary)
	if summary == "" {
		return Record{}, errors.New("summary is required")
	}
	root, dir, err := s.gitLocations()
	if err != nil {
		return Record{}, err
	}
	commit, err := git(root, "rev-parse", "--verify", "HEAD^{commit}")
	if err != nil {
		return Record{}, err
	}
	changed, err := git(root, "diff-tree", "--root", "--no-commit-id", "--name-only", "-z", "-r", commit)
	if err != nil {
		return Record{}, err
	}
	files := []string{}
	for _, file := range strings.Split(changed, "\x00") {
		if file != "" {
			files = append(files, file)
		}
	}
	var idBytes [16]byte
	if _, err := rand.Read(idBytes[:]); err != nil {
		return Record{}, err
	}
	r := Record{Version: 1, ID: hex.EncodeToString(idBytes[:]), Repository: root, Commit: commit, SessionID: strings.TrimSpace(sessionID), TaskID: strings.TrimSpace(taskID), Summary: summary, Verification: verification, Files: files, CreatedAt: time.Now().UTC()}
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return Record{}, err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return Record{}, err
	}
	name := filepath.Join(dir, r.ID+".json")
	f, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return Record{}, err
	}
	defer f.Close()
	if _, err = f.Write(append(data, '\n')); err != nil {
		os.Remove(name)
		return Record{}, err
	}
	if err = f.Sync(); err != nil {
		os.Remove(name)
		return Record{}, err
	}
	return r, nil
}

func (s Store) List(commit string) ([]Record, error) {
	_, dir, err := s.gitLocations()
	if err != nil {
		return nil, err
	}
	if commit != "" {
		root, _, _ := s.gitLocations()
		commit, err = git(root, "rev-parse", "--verify", commit+"^{commit}")
		if err != nil {
			return nil, err
		}
	}
	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return []Record{}, nil
	}
	if err != nil {
		return nil, err
	}
	var out []Record
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		id := strings.TrimSuffix(entry.Name(), ".json")
		if !validID(id) {
			continue
		}
		r, err := s.Get(id)
		if err != nil {
			return nil, err
		}
		if commit == "" || r.Commit == commit {
			out = append(out, r)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

func (s Store) Get(id string) (Record, error) {
	if !validID(id) {
		return Record{}, errors.New("invalid checkpoint ID")
	}
	_, dir, err := s.gitLocations()
	if err != nil {
		return Record{}, err
	}
	data, err := os.ReadFile(filepath.Join(dir, id+".json"))
	if err != nil {
		return Record{}, err
	}
	var r Record
	if err := json.Unmarshal(data, &r); err != nil {
		return Record{}, err
	}
	if r.Version != 1 || r.ID != id {
		return Record{}, errors.New("unsupported or mismatched checkpoint")
	}
	return r, nil
}

func (s Store) gitLocations() (string, string, error) {
	repo := s.Repo
	if repo == "" {
		repo = "."
	}
	root, err := git(repo, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", "", err
	}
	common, err := git(root, "rev-parse", "--git-common-dir")
	if err != nil {
		return "", "", err
	}
	if !filepath.IsAbs(common) {
		common = filepath.Join(root, common)
	}
	return root, filepath.Join(common, "record", "checkpoints"), nil
}

func git(repo string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", repo}, args...)...)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return strings.TrimSuffix(string(out), "\n"), nil
}

func validID(id string) bool {
	if len(id) != 32 {
		return false
	}
	_, err := hex.DecodeString(id)
	return err == nil && strings.ToLower(id) == id
}
