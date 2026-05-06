package main

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"
	"time"
)

// ============================================
// Content-Addressable Storage (CAS) + Merkle DAG
// ============================================
//
// This file implements a tiny, Git-inspired storage model:
//
//   - Blob   = raw file bytes
//   - Tree   = a directory: name -> object hash (usually blob hashes)
//   - Commit = a snapshot: points to one root tree + one parent commit
//
// The key idea is *content addressing*: object IDs are hashes of their content.
// That gives you:
//   - Integrity: if bytes change, the hash changes.
//   - Deduplication: identical content is stored once and referenced many times.
//   - Merkle DAG: commits/trees/blobs form a directed acyclic graph of hashes.

// Hash is a SHA-256 digest. In this demo we use SHA-256 for familiarity; Git uses
// SHA-1 historically and can also use SHA-256 in newer modes.
type Hash [32]byte

func (h Hash) String() string { return fmt.Sprintf("%x", h[:4]) }
func (h Hash) Full() string   { return fmt.Sprintf("%x", h[:]) }

func hashData(data []byte) Hash { return sha256.Sum256(data) }

// CAS stores objects keyed by their hash (their "content address").
//
// Important: for CAS to be meaningful, objects must behave immutably after being
// written. This implementation enforces that by copying bytes/maps on write and
// on read.
type CAS struct {
	blobs   map[Hash][]byte // Raw file content
	trees   map[Hash]Tree   // Directory structures
	commits map[Hash]Commit // Version snapshots
}

func NewCAS() *CAS {
	return &CAS{
		blobs:   make(map[Hash][]byte),
		trees:   make(map[Hash]Tree),
		commits: make(map[Hash]Commit),
	}
}

// ============================================
// Blob: raw file content
// ============================================

type Blob struct {
	Content []byte
}

func (c *CAS) WriteBlob(content []byte) Hash {
	// Copy so callers can't mutate our stored data later.
	contentCopy := append([]byte(nil), content...)
	hash := hashData(contentCopy)

	if _, exists := c.blobs[hash]; !exists {
		c.blobs[hash] = contentCopy
		fmt.Printf("  [NEW BLOB] %s (%d bytes)\n", hash, len(contentCopy))
	} else {
		fmt.Printf("  [REUSED BLOB] %s\n", hash)
	}
	return hash
}

func (c *CAS) ReadBlob(hash Hash) ([]byte, bool) {
	content, ok := c.blobs[hash]
	if !ok {
		return nil, false
	}
	// Return a copy to preserve immutability semantics.
	return append([]byte(nil), content...), true
}

// ============================================
// Tree: directory (name -> blob/tree hash)
// ============================================

type Tree struct {
	Entries map[string]Hash // filename or dirname -> hash
}

func (c *CAS) WriteTree(entries map[string]Hash) Hash {
	// NOTE: In Git, a "tree" object encodes entries as:
	//   "<mode> <name>\x00<20-byte SHA1>"
	// We simplify that to:
	//   "<name>\x00<32-byte SHA256>"
	//
	// Two important properties:
	//   1) Determinism: sort names so the same directory hashes the same.
	//   2) Structural hashing: change any child => tree hash changes.

	names := make([]string, 0, len(entries))
	for name := range entries {
		names = append(names, name)
	}
	sort.Strings(names)

	estimatedSize := 0
	for _, name := range names {
		estimatedSize += len(name) + 1 + len(Hash{}) // name + NUL + hash
	}
	treeData := make([]byte, 0, estimatedSize)
	for _, name := range names {
		treeData = append(treeData, name...)
		treeData = append(treeData, 0)
		h := entries[name]
		treeData = append(treeData, h[:]...)
	}

	hash := hashData(treeData)
	if _, exists := c.trees[hash]; !exists {
		entriesCopy := make(map[string]Hash, len(entries))
		for k, v := range entries {
			entriesCopy[k] = v
		}
		c.trees[hash] = Tree{Entries: entriesCopy}
		fmt.Printf("  [NEW TREE] %s (%d entries)\n", hash, len(entriesCopy))
	}
	return hash
}

func (c *CAS) ReadTree(hash Hash) (Tree, bool) {
	tree, ok := c.trees[hash]
	if !ok {
		return Tree{}, false
	}
	entriesCopy := make(map[string]Hash, len(tree.Entries))
	for k, v := range tree.Entries {
		entriesCopy[k] = v
	}
	return Tree{Entries: entriesCopy}, true
}

// ============================================
// Commit: version snapshot
// ============================================

type Commit struct {
	TreeHash  Hash
	Parent    Hash // Previous commit (zero hash for first)
	Author    string
	Message   string
	Timestamp time.Time
}

func (c *CAS) WriteCommit(treeHash Hash, parent Hash, author, message string) Hash {
	// A commit hash depends on:
	//   - which tree it points to (snapshot content)
	//   - which parent it points to (history)
	//   - metadata (author/message/time)
	now := time.Now().UTC()
	commitData := []byte(fmt.Sprintf("%s\n%s\n%s\n%s\n%s",
		treeHash.Full(),
		parent.Full(),
		author,
		message,
		now.Format(time.RFC3339),
	))

	hash := hashData(commitData)
	if _, exists := c.commits[hash]; !exists {
		c.commits[hash] = Commit{
			TreeHash:  treeHash,
			Parent:    parent,
			Author:    author,
			Message:   message,
			Timestamp: now,
		}
		fmt.Printf("  [NEW COMMIT] %s: %s\n", hash, message)
	}
	return hash
}

func (c *CAS) ReadCommit(hash Hash) (Commit, bool) {
	commit, ok := c.commits[hash]
	return commit, ok
}

// ============================================
// Repository: high-level operations
// ============================================

type Repository struct {
	cas    *CAS
	head   Hash // Tip of the current branch (zero hash for "no commits yet")
	branch string
}

func NewRepository() *Repository {
	return &Repository{
		cas:    NewCAS(),
		head:   Hash{},
		branch: "main",
	}
}

// WriteFiles creates blobs and a root tree from a file map.
//
// For simplicity, all files are treated as being in the repository root.
// Real Git recursively builds trees for nested directories.
func (r *Repository) WriteFiles(files map[string][]byte) Hash {
	entries := make(map[string]Hash, len(files))
	for path, content := range files {
		blobHash := r.cas.WriteBlob(content)
		entries[path] = blobHash
	}
	return r.cas.WriteTree(entries)
}

// Commit creates a new snapshot and advances HEAD.
func (r *Repository) Commit(files map[string][]byte, message string) Hash {
	treeHash := r.WriteFiles(files)
	commitHash := r.cas.WriteCommit(treeHash, r.head, "You", message)
	r.head = commitHash
	return commitHash
}

// Restore materializes all files from a commit's root tree.
func (r *Repository) Restore(commitHash Hash) map[string][]byte {
	commit, ok := r.cas.ReadCommit(commitHash)
	if !ok {
		return nil
	}
	tree, ok := r.cas.ReadTree(commit.TreeHash)
	if !ok {
		return nil
	}

	files := make(map[string][]byte, len(tree.Entries))
	for name, blobHash := range tree.Entries {
		content, ok := r.cas.ReadBlob(blobHash)
		if ok {
			files[name] = content
		}
	}
	return files
}

// History walks parent pointers back to the root commit.
func (r *Repository) History(start Hash) []Hash {
	history := []Hash{}
	current := start
	for current != (Hash{}) {
		history = append(history, current)
		commit, ok := r.cas.ReadCommit(current)
		if !ok {
			break
		}
		current = commit.Parent
	}
	return history
}

func (r *Repository) ShowCommit(hash Hash) {
	commit, ok := r.cas.ReadCommit(hash)
	if !ok {
		fmt.Printf("Commit %s not found\n", hash)
		return
	}

	fmt.Println()
	fmt.Println("------------------------------------------------------------")
	fmt.Printf("Commit:  %s\n", hash)
	fmt.Printf("Author:  %s\n", commit.Author)
	fmt.Printf("Date:    %s\n", commit.Timestamp.Format("2006-01-02 15:04:05 MST"))
	fmt.Printf("Message: %s\n", commit.Message)
	fmt.Printf("Tree:    %s\n", commit.TreeHash)
	if commit.Parent != (Hash{}) {
		fmt.Printf("Parent:  %s\n", commit.Parent)
	}
	fmt.Println("------------------------------------------------------------")
}

// ShowDAG prints the Merkle DAG (starting from a commit and walking parents).
//
// This is a "graph of hashes": commits point to trees, trees point to blobs.
func (r *Repository) ShowDAG(commitHash Hash, indent string, visited map[Hash]bool) {
	if visited[commitHash] {
		fmt.Printf("%s[seen] Commit %s\n", indent, commitHash)
		return
	}
	visited[commitHash] = true

	commit, ok := r.cas.ReadCommit(commitHash)
	if !ok {
		return
	}

	fmt.Printf("%sCommit %s: %s\n", indent, commitHash, commit.Message)

	tree, ok := r.cas.ReadTree(commit.TreeHash)
	if ok {
		fmt.Printf("%s  Tree %s\n", indent, commit.TreeHash)
		names := make([]string, 0, len(tree.Entries))
		for name := range tree.Entries {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			blobHash := tree.Entries[name]
			fmt.Printf("%s    %s -> %s\n", indent, name, blobHash)
		}
	}

	if commit.Parent != (Hash{}) {
		fmt.Printf("%s  Parent: %s\n", indent, commit.Parent)
		r.ShowDAG(commit.Parent, indent+"    ", visited)
	}
}

func (r *Repository) Stats() {
	fmt.Printf("\n=== Storage Statistics ===\n")
	fmt.Printf("Unique blobs:   %d\n", len(r.cas.blobs))
	fmt.Printf("Unique trees:   %d\n", len(r.cas.trees))
	fmt.Printf("Unique commits: %d\n", len(r.cas.commits))

	totalBlobBytes := 0
	for _, content := range r.cas.blobs {
		totalBlobBytes += len(content)
	}
	fmt.Printf("Total blob data: %d bytes\n", totalBlobBytes)
}

// ============================================
// Demo
// ============================================

func main() {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("GIT-STYLE CAS WITH MERKLE DAG (Learning Demo)")
	fmt.Println(strings.Repeat("=", 60))

	repo := NewRepository()

	fmt.Println("\nVERSION 1: Initial commit")
	files1 := map[string][]byte{
		"README.md":   []byte("# My Project\nFirst version"),
		"main.go":     []byte("package main\n\nfunc main() {\n    println(\"v1\")\n}\n"),
		"config.json": []byte("{\"version\": 1}"),
	}
	commit1 := repo.Commit(files1, "Initial commit")
	fmt.Printf("\nOK: Commit 1 created: %s\n", commit1)

	fmt.Println("\nVERSION 2: Update README.md only")
	files2 := map[string][]byte{
		"README.md":   []byte("# My Project\nSecond version - added docs"),
		"main.go":     files1["main.go"],     // same as v1
		"config.json": files1["config.json"], // same as v1
	}
	commit2 := repo.Commit(files2, "Update README")
	fmt.Printf("\nOK: Commit 2 created: %s\n", commit2)

	fmt.Println("\nVERSION 3: Add feature.go and modify main.go")
	files3 := map[string][]byte{
		"README.md":   files2["README.md"], // same as v2
		"main.go":     []byte("package main\n\nfunc main() {\n    println(\"v2\")\n    println(\"feature enabled\")\n}\n"),
		"config.json": files1["config.json"], // same as v1
		"feature.go":  []byte("package main\n\nfunc Feature() {\n    println(\"feature v1\")\n}\n"),
	}
	commit3 := repo.Commit(files3, "Add feature module")
	fmt.Printf("\nOK: Commit 3 created: %s\n", commit3)

	repo.Stats()

	fmt.Println("\n=== COMMIT DETAILS (Pointers) ===")
	repo.ShowCommit(commit1)
	repo.ShowCommit(commit2)
	repo.ShowCommit(commit3)

	fmt.Println("\n=== MERKLE DAG STRUCTURE ===")
	visited := make(map[Hash]bool)
	repo.ShowDAG(commit3, "", visited)

	fmt.Println("\n=== RESTORING PAST VERSIONS ===")

	fmt.Println("\nRestoring Version 1:")
	v1Files := repo.Restore(commit1)
	for name, content := range v1Files {
		fmt.Printf("  %s: %q\n", name, content)
	}

	fmt.Println("\nRestoring Version 2:")
	v2Files := repo.Restore(commit2)
	for name, content := range v2Files {
		fmt.Printf("  %s: %q\n", name, content)
	}

	fmt.Println("\nRestoring Version 3:")
	v3Files := repo.Restore(commit3)
	for name, content := range v3Files {
		fmt.Printf("  %s: %q\n", name, content[:min(50, len(content))])
	}

	fmt.Println("\n=== HISTORY (Walking Parent Pointers) ===")
	history := repo.History(commit3)
	for i, h := range history {
		commit, _ := repo.cas.ReadCommit(h)
		fmt.Printf("%d. %s: %s\n", i+1, h, commit.Message)
	}

	fmt.Println("\n=== DEDUPLICATION PROOF ===")
	fmt.Println("\nEven though we have 3 versions, notice:")
	fmt.Println("  - main.go from v1 and v2 are the same blob")
	fmt.Println("  - config.json is the same across all versions")
	fmt.Println("  - README.md from v2 and v3 are the same blob")
	fmt.Println("\nEach unique content is stored only once.")
	fmt.Println("Versions just POINT to existing blobs via hashes.")

	fmt.Println("\nGit-style CAS with Merkle DAG complete.")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
