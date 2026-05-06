package main

import (
	"bytes"
	"testing"
)

func TestDeduplicationAndRestore(t *testing.T) {
	repo := NewRepository()

	files1 := map[string][]byte{
		"README.md":   []byte("# My Project\nFirst version"),
		"main.go":     []byte("package main\n\nfunc main() {\n    println(\"v1\")\n}\n"),
		"config.json": []byte("{\"version\": 1}"),
	}
	commit1 := repo.Commit(files1, "Initial commit")

	files2 := map[string][]byte{
		"README.md":   []byte("# My Project\nSecond version - added docs"),
		"main.go":     files1["main.go"],     // same as v1
		"config.json": files1["config.json"], // same as v1
	}
	commit2 := repo.Commit(files2, "Update README")

	files3 := map[string][]byte{
		"README.md":   files2["README.md"], // same as v2
		"main.go":     []byte("package main\n\nfunc main() {\n    println(\"v2\")\n    println(\"feature enabled\")\n}\n"),
		"config.json": files1["config.json"], // same as v1
		"feature.go":  []byte("package main\n\nfunc Feature() {\n    println(\"feature v1\")\n}\n"),
	}
	commit3 := repo.Commit(files3, "Add feature module")

	// Unique blobs:
	// - v1: README, main, config (3)
	// - v2 adds a new README (1 more => 4)
	// - v3 adds a new main + feature (2 more => 6)
	if got, want := len(repo.cas.blobs), 6; got != want {
		t.Fatalf("unique blobs = %d, want %d", got, want)
	}
	if got, want := len(repo.cas.trees), 3; got != want {
		t.Fatalf("unique trees = %d, want %d", got, want)
	}
	if got, want := len(repo.cas.commits), 3; got != want {
		t.Fatalf("unique commits = %d, want %d", got, want)
	}

	restore1 := repo.Restore(commit1)
	if !bytes.Equal(restore1["README.md"], files1["README.md"]) {
		t.Fatalf("restore v1 README mismatch")
	}
	if !bytes.Equal(restore1["main.go"], files1["main.go"]) {
		t.Fatalf("restore v1 main.go mismatch")
	}
	if !bytes.Equal(restore1["config.json"], files1["config.json"]) {
		t.Fatalf("restore v1 config.json mismatch")
	}

	restore2 := repo.Restore(commit2)
	if !bytes.Equal(restore2["README.md"], files2["README.md"]) {
		t.Fatalf("restore v2 README mismatch")
	}

	restore3 := repo.Restore(commit3)
	if !bytes.Equal(restore3["feature.go"], files3["feature.go"]) {
		t.Fatalf("restore v3 feature.go mismatch")
	}

	// A sanity check that parent pointers link the commits (history length).
	if got, want := len(repo.History(commit3)), 3; got != want {
		t.Fatalf("history length = %d, want %d", got, want)
	}

	// Immutability check: mutating a caller-owned slice should not affect stored data.
	mut := repo.Restore(commit1)
	mut["README.md"][0] = 'X'
	again := repo.Restore(commit1)
	if bytes.Equal(mut["README.md"], again["README.md"]) {
		t.Fatalf("expected restored content copies; mutation leaked into storage")
	}
}
