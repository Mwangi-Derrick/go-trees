# Exploring Trees in Go: Core Foundations

A deep dive into tree-based data structures, starting with the implementation of **Merkle Trees** for data integrity. This repository serves as a foundational base for understanding how trees underpin modern systems—from databases and storage engines to distributed systems and file systems.

## 🎓 Learning Objective

The goal of this project is to explore the "why" and "how" behind specialized tree structures in Go. By building these from the ground up, we examine how memory layout, pointer management, and recursive logic can be optimized for real-world applications.

### Why Trees?
Trees are the silent workhorses of infrastructure:
- **Databases:** Use B-Trees and LSM Trees for indexing and fast retrieval.
- **Storage & File Systems:** Use Merkle Trees and Radix Trees for deduplication and integrity.
- **Distributed Systems:** Use Merkle Trees for efficient state synchronization (e.g., in blockchain or P2P networks).

## 🌲 Current Implementation: Merkle Trees

This repo begins with a Go-based implementation of a **Merkle Tree**, a specialized structure where every node is a cryptographic hash of its children.

### Core Concepts Explored
- **Hash Pointers:** Moving beyond memory addresses to cryptographic commitments.
- **Integrity Verification:** Learning how a single root hash can verify gigabytes of data.
- **Tree Traversal & Mutation:** Managing parent-child relationships and recomputing state on updates.
- **Package Organization:** Separating node definitions (`data/`) from tree logic to simulate real-world library structures.

## 🛠️ Project Structure

```text
.
├── merkle.go         # Main engine logic and tree construction
├── go.mod            # Go module definition
└── data/             # Domain-specific node definitions
    ├── merkle_node.go # Structure for hash-based nodes
    └── normal_node.go # Standard data nodes for comparison
```

## 🚦 Getting Started

### Prerequisites
- Go 1.18+

### Installation & Run
```bash
git clone <repo-url>
cd merkle
go run merkle.go
```

## 🚀 Future Explorations

This repository is designed to evolve as a playground for more advanced structures:
- [ ] **B-Trees & B+ Trees:** Understanding disk-aware indexing for database engines.
- [ ] **LSM Trees (Log-Structured Merge-Trees):** Exploring write-optimized storage.
- [ ] **Self-Balancing Trees:** Implementing AVL or Red-Black logic to prevent degenerate "chain" states.
- [ ] **Merkle Proofs:** Implementing the logic for efficient inclusion proofs.

---

*Built as a foundational study in Go Systems Engineering.*

## Running the Demos

Each top-level folder is a small, self-contained learning demo (most are `package main`).

- Merkle DAG (Git-style CAS): `go run ./merkle-DAG/git-dag.go`
- LSM Tree: `go run ./LSM/lsm.go`
