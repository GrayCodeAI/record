package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/GrayCodeAI/record/checkpoint"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "record:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return errors.New("use create, list, or show")
	}
	cmd := args[0]
	fs := flag.NewFlagSet(cmd, flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	repo := fs.String("repo", ".", "Git repository or worktree")
	session := fs.String("session", "", "agent session ID")
	task := fs.String("task", "", "work item ID")
	summary := fs.String("summary", "", "checkpoint summary")
	verification := fs.String("verification", "", "verification result")
	commit := fs.String("commit", "", "Git commit or ref")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	store := checkpoint.Store{Repo: *repo}
	switch cmd {
	case "create":
		var checks []string
		if strings.TrimSpace(*verification) != "" {
			checks = []string{strings.TrimSpace(*verification)}
		}
		r, err := store.Create(*session, *task, *summary, checks)
		if err != nil {
			return err
		}
		return printJSON(r)
	case "list":
		r, err := store.List(*commit)
		if err != nil {
			return err
		}
		return printJSON(r)
	case "show":
		if fs.NArg() == 1 {
			r, err := store.Get(fs.Arg(0))
			if err != nil {
				return err
			}
			return printJSON(r)
		}
		if *commit == "" {
			return errors.New("show requires a checkpoint ID or --commit <ref>")
		}
		r, err := store.List(*commit)
		if err != nil {
			return err
		}
		return printJSON(r)
	default:
		return fmt.Errorf("unknown command %q", cmd)
	}
}

func printJSON(v any) error { return json.NewEncoder(os.Stdout).Encode(v) }
