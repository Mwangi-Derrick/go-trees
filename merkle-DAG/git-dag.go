package main

import (
   // "bytes"
    "crypto/sha256"
    "fmt"
    "sort"
    "strings"
    "time"
)

// ============================================
// Content-Addressable Storage
// ============================================

type Hash [32]byte

func (h Hash) String() string {
    return fmt.Sprintf("%x", h[:4])
}

func (h Hash) Full() string {
    return fmt.Sprintf("%x", h[:])
}

func hashData(data []byte) Hash {
    return sha256.Sum256(data)
}

type CAS struct {
    blobs   map[Hash][]byte   // Raw file content
    trees   map[Hash]Tree     // Directory structures
    commits map[Hash]Commit   // Version snapshots
}

func NewCAS() *CAS {
    return &CAS{
        blobs:   make(map[Hash][]byte),
        trees:   make(map[Hash]Tree),
        commits: make(map[Hash]Commit),
    }
}

// ============================================
// Blob: Raw file content
// ============================================

type Blob struct {
    Content []byte
}

func (c *CAS) WriteBlob(content []byte) Hash {
    hash := hashData(content)
    if _, exists := c.blobs[hash]; !exists {
        c.blobs[hash] = content
        fmt.Printf("  📄 NEW BLOB: %s (%d bytes)\n", hash, len(content))
    } else {
        fmt.Printf("  ♻️ REUSED BLOB: %s (already exists)\n", hash)
    }
    return hash
}

func (c *CAS) ReadBlob(hash Hash) ([]byte, bool) {
    content, ok := c.blobs[hash]
    return content, ok
}

// ============================================
// Tree: Directory (filename → blob/tree hash)
// ============================================

type Tree struct {
    Entries map[string]Hash  // filename or dirname → hash
}

func (c *CAS) WriteTree(entries map[string]Hash) Hash {
    // Sort entries for deterministic hashing
    names := make([]string, 0, len(entries))
    for name := range entries {
        names = append(names, name)
    }
    sort.Strings(names)
    
	var treeData []byte
	for _, name := range names {
		treeData = append(treeData, []byte(name+"\x00")...)
		h := entries[name]
		treeData = append(treeData, h[:]...)
	}
    
    hash := hashData(treeData)
    if _, exists := c.trees[hash]; !exists {
        c.trees[hash] = Tree{Entries: entries}
        fmt.Printf("  📁 NEW TREE: %s (%d entries)\n", hash, len(entries))
    }
    return hash
}

func (c *CAS) ReadTree(hash Hash) (Tree, bool) {
    tree, ok := c.trees[hash]
    return tree, ok
}

// ============================================
// Commit: Version snapshot
// ============================================

type Commit struct {
    TreeHash  Hash
    Parent    Hash      // Previous commit (zero hash for first)
    Author    string
    Message   string
    Timestamp time.Time
}

func (c *CAS) WriteCommit(treeHash Hash, parent Hash, author, message string) Hash {
    commitData := []byte(fmt.Sprintf("%s\n%s\n%s\n%s\n%s",
        treeHash.Full(),
        parent.Full(),
        author,
        message,
        time.Now().Format(time.RFC3339),
    ))
    
    hash := hashData(commitData)
    if _, exists := c.commits[hash]; !exists {
        c.commits[hash] = Commit{
            TreeHash:  treeHash,
            Parent:    parent,
            Author:    author,
            Message:   message,
            Timestamp: time.Now(),
        }
        fmt.Printf("  ✨ NEW COMMIT: %s: %s\n", hash, message)
    }
    return hash
}

func (c *CAS) ReadCommit(hash Hash) (Commit, bool) {
    commit, ok := c.commits[hash]
    return commit, ok
}

// ============================================
// Repository: High-level operations
// ============================================

type Repository struct {
    cas   *CAS
    head  Hash  // Current commit
    branch string
}

func NewRepository() *Repository {
    return &Repository{
        cas:   NewCAS(),
        head:  Hash{},
        branch: "main",
    }
}

// WriteFiles creates blobs and tree from a file map
func (r *Repository) WriteFiles(files map[string][]byte) Hash {
    entries := make(map[string]Hash)
    
    for path, content := range files {
        // For simplicity, assume all files are in root directory
        // (Real Git handles nested directories recursively)
        blobHash := r.cas.WriteBlob(content)
        entries[path] = blobHash
    }
    
    return r.cas.WriteTree(entries)
}

// Commit creates a new version
func (r *Repository) Commit(files map[string][]byte, message string) Hash {
    treeHash := r.WriteFiles(files)
    commitHash := r.cas.WriteCommit(treeHash, r.head, "You", message)
    r.head = commitHash
    return commitHash
}

// Restore gets all files from a commit
func (r *Repository) Restore(commitHash Hash) map[string][]byte {
    commit, ok := r.cas.ReadCommit(commitHash)
    if !ok {
        return nil
    }
    
    tree, ok := r.cas.ReadTree(commit.TreeHash)
    if !ok {
        return nil
    }
    
    files := make(map[string][]byte)
    for name, blobHash := range tree.Entries {
        content, ok := r.cas.ReadBlob(blobHash)
        if ok {
            files[name] = content
        }
    }
    return files
}

// History walks back through parent pointers
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

// ShowCommit prints commit details
func (r *Repository) ShowCommit(hash Hash) {
    commit, ok := r.cas.ReadCommit(hash)
    if !ok {
        fmt.Printf("Commit %s not found\n", hash)
        return
    }
    
    fmt.Printf("\n┌─────────────────────────────────────────┐\n")
    fmt.Printf("│ Commit: %s\n", hash)
    fmt.Printf("│ Author: %s\n", commit.Author)
    fmt.Printf("│ Date:   %s\n", commit.Timestamp.Format("2006-01-02 15:04:05"))
    fmt.Printf("│ Message: %s\n", commit.Message)
    fmt.Printf("│ Tree:   %s\n", commit.TreeHash)
    if commit.Parent != (Hash{}) {
        fmt.Printf("│ Parent: %s\n", commit.Parent)
    }
    fmt.Printf("└─────────────────────────────────────────┘")
}

// ShowDAG prints the Merkle DAG structure
func (r *Repository) ShowDAG(commitHash Hash, indent string, visited map[Hash]bool) {
    if visited[commitHash] {
        fmt.Printf("%s♻️ Commit %s (already shown)\n", indent, commitHash)
        return
    }
    visited[commitHash] = true
    
    commit, ok := r.cas.ReadCommit(commitHash)
    if !ok {
        return
    }
    
    fmt.Printf("%s📦 Commit %s: %s\n", indent, commitHash, commit.Message)
    
    // Show tree
    tree, ok := r.cas.ReadTree(commit.TreeHash)
    if ok {
        fmt.Printf("%s  📁 Tree %s\n", indent, commit.TreeHash)
        for name, blobHash := range tree.Entries {
            fmt.Printf("%s    📄 %s → %s\n", indent, name, blobHash)
        }
    }
    
    // Show parent
    if commit.Parent != (Hash{}) {
        fmt.Printf("%s  └─ Parent: %s\n", indent, commit.Parent)
        r.ShowDAG(commit.Parent, indent+"     ", visited)
    }
}

// Stats shows storage efficiency
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
    fmt.Println("GIT-STYLE CAS WITH MERKLE DAG")
    fmt.Println(strings.Repeat("=", 60))
    
    repo := NewRepository()
    
    // ========================================
    // Version 1: Initial commit
    // ========================================
    fmt.Println("\n🔹 VERSION 1: Initial commit")
    files1 := map[string][]byte{
        "README.md": []byte("# My Project\nFirst version"),
        "main.go":   []byte(`package main

func main() {
    println("v1")
}`),
        "config.json": []byte(`{"version": 1}`),
    }
    
    commit1 := repo.Commit(files1, "Initial commit")
    fmt.Printf("\n✅ Commit 1 created: %s\n", commit1)
    
    // ========================================
    // Version 2: Only README.md changes
    // ========================================
    fmt.Println("\n🔹 VERSION 2: Update README.md only")
    files2 := map[string][]byte{
        "README.md": []byte("# My Project\nSecond version - added docs"),
        "main.go":   files1["main.go"],      // SAME as v1
        "config.json": files1["config.json"], // SAME as v1
    }
    
    commit2 := repo.Commit(files2, "Update README")
    fmt.Printf("\n✅ Commit 2 created: %s\n", commit2)
    
    // ========================================
    // Version 3: Add new file, modify main.go
    // ========================================
    fmt.Println("\n🔹 VERSION 3: Add feature.go and modify main.go")
    files3 := map[string][]byte{
        "README.md":   files2["README.md"],  // SAME as v2
        "main.go":     []byte(`package main

func main() {
    println("v2")
    println("feature enabled")
}`), // CHANGED
        "config.json": files1["config.json"], // SAME as v1
        "feature.go":  []byte(`package main

func Feature() {
    println("feature v1")
}`), // NEW
    }
    
    commit3 := repo.Commit(files3, "Add feature module")
    fmt.Printf("\n✅ Commit 3 created: %s\n", commit3)
    
    // ========================================
    // Show storage deduplication
    // ========================================
    repo.Stats()
    
    // ========================================
    // Show each commit's pointers
    // ========================================
    fmt.Println("\n=== COMMIT DETAILS (Showing Pointers) ===")
    repo.ShowCommit(commit1)
    repo.ShowCommit(commit2)
    repo.ShowCommit(commit3)
    
    // ========================================
    // Show the Merkle DAG
    // ========================================
    fmt.Println("\n=== MERKLE DAG STRUCTURE ===")
    visited := make(map[Hash]bool)
    repo.ShowDAG(commit3, "", visited)
    
    // ========================================
    // Demonstrate restoration
    // ========================================
    fmt.Println("\n=== RESTORING PAST VERSIONS ===")
    
    fmt.Println("\n📂 Restoring Version 1:")
    v1Files := repo.Restore(commit1)
    for name, content := range v1Files {
        fmt.Printf("  %s: %q\n", name, content)
    }
    
    fmt.Println("\n📂 Restoring Version 2:")
    v2Files := repo.Restore(commit2)
    for name, content := range v2Files {
        fmt.Printf("  %s: %q\n", name, content)
    }
    
    fmt.Println("\n📂 Restoring Version 3:")
    v3Files := repo.Restore(commit3)
    for name, content := range v3Files {
        fmt.Printf("  %s: %q\n", name, content[:min(50, len(content))])
    }
    
    // ========================================
    // Show history (walking parent pointers)
    // ========================================
    fmt.Println("\n=== HISTORY (Walking Parent Pointers) ===")
    history := repo.History(commit3)
    for i, h := range history {
        commit, _ := repo.cas.ReadCommit(h)
        fmt.Printf("%d. %s: %s\n", i+1, h, commit.Message)
    }
    
    // ========================================
    // Deduplication proof
    // ========================================
    fmt.Println("\n=== DEDUPLICATION PROOF ===")
    fmt.Println("\nEven though we have 3 versions, notice:")
    fmt.Println("  • main.go from v1 and v2 are the SAME blob")
    fmt.Println("  • config.json is the SAME across all versions")
    fmt.Println("  • README.md from v2 and v3 are the SAME blob")
    fmt.Println("\nEach unique content is stored only ONCE.")
    fmt.Println("Versions just POINT to existing blobs via hashes.")
    
    fmt.Println("\n Git-style CAS with Merkle DAG complete!")
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}